# async-component
Async component for knative services

![diagram](./README-images/diagram.png)

1. Follow instructions for installing Kafka Source, but do not create event display service (https://knative.dev/docs/eventing/samples/kafka/source/)

1. Apply config files:
    ```
    ko apply -f config/async-requests
    ```
1. Make note of your Kubernetes service external IP.
    ```
    kubectl get service producer-service
    ```
1. For now, modify /etc/hosts to point traffic from your application (something like myapp.default.11.112.113.14) your Kubernetes service IP (something like 11.111.111.11)
    ```
    11.111.111.11   myapp.default.11.112.113.14.xip.io
    ```

1. Curl your application. Try async & non async.

    ```
    curl myapp.default.11.112.113.14.xip.io
    curl myapp.default.11.112.113.14.xip.io -H "Prefer: respond-async" -v
    ```

1. Operator found here:
kubectl create -f https://raw.githubusercontent.com/spotahome/redis-operator/master/example/operator/all-redis-operator-resources.yaml
