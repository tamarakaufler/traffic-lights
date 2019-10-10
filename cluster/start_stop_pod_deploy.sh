#!/bin/bash
echo "kubectl apply -f traffic_lights_deploy.yaml"
echo "then 10 rounds of pod deletion"
echo ""
echo "Start ..."
sleep 5
kubectl apply -f traffic_lights_deploy.yaml
sleep 3

for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20
do
    echo "round $i"
    echo "creating pod at `date`"
    sleep 3
    echo "deleting pod at `date`"
    kubectl delete pod -l 'app=traffic-lights'
done
kubectl delete deployments.apps/traffic-lights

echo "End ..."


