# async-component
Async component for knative services

1. Follow instructions for installing Kafka Source, but do not create event display service (https://knative.dev/docs/eventing/samples/kafka/source/)
2. Apply config files:
    ```
    kubectl apply -f config/async-requests
    ```
3. 