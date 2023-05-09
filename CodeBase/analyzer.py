import matplotlib as mpl
from scipy.stats import binom 
import os
import pickle


main_Dict = {}

empty_event = {"eventno" : 0, "eType" : "", "eSend" : 0, "eRecv" : -1, "eInt" : False, "eVclock" : [], "eBclock" : []}

def getfilename(foldername):
    file_names = filter((lambda x:x[-3:] == 'txt'),os.listdir(os.getcwd()+"/"+foldername))
    return sorted(list(file_names))

def get_event(eventLinearr,tempVclock,tempBclock):
    result = {}

    result["eventno"] = int(eventLinearr[0])
    result["eType"] = eventLinearr[1]
    result["eSend"] = int(eventLinearr[2])
    if result["eType"] == "INTERNAL EVENT":
        result["eInt"] = True
    else:
        result["eRecv"] = int(eventLinearr[3])
        result["eInt"] = False

    result["eVclock"] = get_vector(tempVclock)
    result["eBclock"] = get_vector(tempBclock)

    return result

def get_vector(vectorline):
    tempLine = vectorline[1:-1]
    vectorClock = list(map((lambda x:int(x)),tempLine.split(" ")))

    return vectorClock


def make_dict(file_name):
    with open(file_name,"r") as fin:
        data = fin.read().split("\n")

    totalGSN = int(data[-3].split(" ")[-2])
    lineNo = 0
    currEvent = 1
    eventList = [None] * (totalGSN+1)
    eventList[0] = empty_event

    while True:
        tempLine = data[lineNo]
        tempVclock = data[lineNo+1]
        tempBclock = data[lineNo+2]

        eventLine = tempLine.split(",")
        if currEvent != int(eventLine[0]):
            print("PANIC",file_name,currEvent,lineNo)
            break

        tempEvent = get_event(eventLine,tempVclock,tempBclock)
        eventList[currEvent] = tempEvent

        if currEvent == totalGSN:
            break

        currEvent += 1
        lineNo += 3

    return eventList    



def get_file_prp(file_name):
    eventList = main_Dict[file_name]

    n = int(file_name[1:4])
    m = n * float(file_name[5:8])
    baseEvent = eventList[10*n]
    idx = (10*n)+10
    gsnlst = []
    prplst = []
    while idx < ((n*n)+10*n):
        gsnlst.append(idx)
        prplst.append(get_event_prp(baseEvent,eventList[idx],m))
        idx += 10

    return gsnlst,prplst

def get_file_prfp(gsnlst, prplst,eventSlice,n):
    baseEvent = eventSlice[10*n]
    prfplst = []
    for idx,val in enumerate(gsnlst):
        prdelta = int(check_two_events_bloom(baseEvent,eventSlice[val]))
        prfplst.append((1 - prplst[idx]) * prdelta)

    return gsnlst, prfplst


def get_event_prp(beforeEvent,afterEvent,bloomsize):
    ans = 1
    bzSum = sum(afterEvent["eBclock"])
    for idx,val in enumerate(beforeEvent["eBclock"]):
        result = 0
        pSuccess = 1/float(bloomsize)
        for val in range(0,val):
            result += binom.pmf(val, bzSum, pSuccess)

        ans = ans *(1 - result)
    return ans



def check_two_events_vec(beforeEvent,afterEvent):
    flag = True
    if beforeEvent["eVclock"] == afterEvent["eVclock"]:
        return False

    for idx,val in enumerate(beforeEvent["eVclock"]):
        if val > afterEvent["eVclock"][idx]:
            flag = False


    return flag

def check_two_events_bloom(beforeEvent,afterEvent):
    flag = True
    if beforeEvent["eBclock"] == afterEvent["eBclock"]:
        return False
    for idx,val in enumerate(beforeEvent["eBclock"]):
        if val > afterEvent["eBclock"][idx]:
            flag = False
            
    return flag



def get_metrics(eventSlice,n):
    indecies = list(range(10*n,n*n+10*n,100))

    totalTP = 0
    totalTN = 0
    totalFP = 0
    totalFN = 0

    for idx,valx in enumerate(indecies[:-1]):
        for valy in indecies[idx+1:]:
            bflag = check_two_events_bloom(eventSlice[valx],eventSlice[valy])
            if bflag:
                if check_two_events_vec(eventSlice[valx],eventSlice[valy]):
                    # True Positive
                    totalTP += 1
                else:
                    # False Positive
                    totalFP += 1
                # bflag True => Bz >= By Possitive Case
            else:
                # True Negative
                totalTN += 1

    sliceAccuracy = get_accuracy(totalTP,totalTN,totalFP,totalFN)
    slicePrecision = get_precision(totalTP,totalTN,totalFP,totalFN)
    slicefpr = get_fpr(totalTP,totalTN,totalFP,totalFN)

    return sliceAccuracy,slicePrecision,slicefpr


def get_accuracy(tp,tn,fp,fn):
    accuracy = float(tp+tn)/(tp+tn+fp+fn)
    return accuracy

def get_precision(tp,tn,fp,fn):
    precision = float(tp)/(tp+fp)
    return precision

def get_fpr(tp,tn,fp,fn):
    fpr = float(fp)/(tn+fp)
    return fpr
    

def get_file_list(keylist,filelist):
    result_list = []
    for name in filelist:
        flag = True
        for val in keylist:
            if val not in name:
                flag = False
                break
        if flag == True:
            result_list.append(name) 

    return result_list


def get_actual_postive(gsnlst, prplst,eventSlice,n):
    base_event = eventSlice[10*n]
    n_gsnlst = []
    n_postivelst = []
    for idx,val in enumerate(gsnlst):
        if check_two_events_vec(base_event,eventSlice[val]):
            n_gsnlst.append(val)
            n_postivelst.append(prplst[idx])

    return n_gsnlst, n_postivelst

mpl.rcParams['figure.dpi'] = 300
import matplotlib.pyplot as plt

def main():
    foldernames = ['runs0']
    for foldername in foldernames:
        file_names = getfilename(foldername)
        for val in file_names:
            print("Reading File ", val)
            eventList = make_dict(foldername+"/"+val)
            main_Dict[val[:-4]] = eventList
            break

        xAxis,yAxis = get_file_prp(val[:-4])
        xAxis,yAxis = get_file_prfp(xAxis,yAxis,main_Dict[val[:-4]],int(val[1:4]))
        plt.scatter(xAxis,yAxis, label='False Positive Points')
        gsnlst,prfplst = get_actual_postive(xAxis,yAxis,main_Dict[val[:-4]],int(val[1:4]))
        plt.scatter(gsnlst,prfplst, label='Actual Positive Points')
        plt.xlabel("Gsn")
        plt.ylabel("Prfp")
        plt.legend(loc='upper right')
        plt.title("Prfp vs GSN \nProcess ="+val[1:4]+" Bloom Factor ="+val[5:8]+"n Hash Function = "+val[9])
        plt.show()


        print("\n--")
        print(val[:-4])
        print("accuracy","precision","fpr")
        metrics = get_metrics(main_Dict[val[:-4]],int(val[1:4]))
        print(metrics[0],metrics[1],metrics[2])
        print("--\n")


main()
