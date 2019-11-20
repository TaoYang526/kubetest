# kubetest
Testing tool for schedulers on K8s, support K8s default scheduler and YuniKorn scheduler.

## Key Metrics
* scheduled pods: pods that have started to run on kubelet(Decided by PodStatus#StartTime).

## Testing Cases
* Throughput
   - Request 50,000 pods via different schedulers and then record the distributions of scheduled pods, draw results of different schedulers on the same chart.
* Node Fairness
   - Request a certain number of pods via different schedulers and then record the distributions of node usage, draw results on charts separately for different schedulers.

## Testing cases for YuniKorn scheduler
* Queue Fairness
   - Prepare queues with different capacities, request pods with different number or resource for these queues, then record the usage of queues, draw results on the same chart.
