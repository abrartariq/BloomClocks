# The Bloom Clock - Golang Code Base

This repository contains the Golang implementation of the Bloom Clock algorithm, as described in the paper "The bloom clock to characterize causality in distributed systems" by Kshemkalyani et al.

		"The Bloom Clock to Characterize Causality in Distributed Systems" introduces a new algorithm for capturing causal relationships between events in a distributed system. The algorithm uses a Bloom filter to efficiently capture causality and outperforms other approaches in terms of message overhead and space complexity. The paper presents a detailed analysis of the algorithm's theoretical properties and practical applicability, making it a significant contribution to the field of distributed systems.

Code Base Also Contains A python file to do the analysis of the Logs Genrated from CodeBase (analyzer.py)

		This Python code analyzes distributed systems using a bloom clock to characterize causality. It reads event logs from text files, constructs event dictionaries, and calculates performance metrics such as accuracy, precision, and false positive rate. The code uses the matplotlib and scipy libraries for data visualization and statistical analysis. It also defines functions for checking the causality of events using vector clocks and bloom filters.

## Dependencies
- murmur3
- fmt
- hash/crc32
- hash/fnv
- hash/maphash
- math/rand
- os
- strconv
- time

## Usage
The Bloom Clock algorithm can be executed by running the `main` function in the `main.go` file, passing three arguments:

1. `arg1`: Total number of processes.
2. `arg2`: Bloom filter size as a fraction of the total number of processes.
3. `arg3`: Number of hash functions for the Bloom filter.

## Code Structure
The implementation contains the following main components:

### event
Defines the event type, which can be of one of the following types:
- TypeEVENT: MSG typed EVENT
- TypePROCESS: MSG typed PROCESS
- EventINT: INTERNAL
- EventSEND: SEND
- EventRECV: RECV
- EventEND: END

### newprocess
Defines the new process type, which includes process ID and its corresponding channel.

### message
Defines the message type, which includes the message type, an event instance, and a new process instance.

### myhash
Function that returns an array of integers representing the indices of the Bloom filter array to which the event should be added.

### createprocess
Function that creates the processes and defines their behavior, including how they handle internal events, send events, receive events, and end events.

### hypervisor
Function that initializes the system by creating the processes and their corresponding channels.

### seqlogger
Function that logs the events occurring during the execution of the Bloom Clock algorithm.

## Credits
This implementation was developed based on the paper "The bloom clock to characterize causality in distributed systems" by Kshemkalyani et al.

## Link
https://link.springer.com/chapter/10.1007/978-3-030-57811-4_25