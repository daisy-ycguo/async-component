# async-component
Async component for knative services

![diagram](./README-images/diagram.png)

## Follow Instructions for Kafka or Redis
  ### Kafka
  1. Follow instructions for installing Kafka Source, but do not create event display service (https://knative.dev/docs/eventing/samples/kafka/source/)

  ### Redis
  1. Create Redis Operator:
    ```
    kubectl create -f https://raw.githubusercontent.com/spotahome/redis-operator/master/example/operator/all-redis-operator-resources.yaml
    ```

  1. Create Redis Failover:
    ```
    kubectl create -f config/async-requests/redis-failover.yaml
    ```

  1. Create the Redis Source. This is a placeholder source until one is availble from knative/eventing:
    ```
    ko apply -f config/async-requests/100-async-redis-source.yaml
    ```

## Install the Consumer and Producer
  1. Apply the following config files:

      ```
      ko apply -f config/async-requests/100-async-consumer.yaml
      ko apply -f config/async-requests/100-async-producer.yaml
      ko apply vs.yaml
      ```

1. Curl your application. Try async & non async.

    ```
    curl myapp.default.11.112.113.14.xip.io
    curl myapp.default.11.112.113.14.xip.io -H "Prefer: respond-async" -v
    ```


## If not using Istio, or vs.yaml doesn't work for you:
1. Make note of your Kubernetes service external IP.
    ```
    kubectl get service producer-service
    ```

1. For now, modify /etc/hosts to point traffic from your application (something like myapp.default.11.112.113.14) to your Kubernetes service IP (something like 11.111.111.11)
    ```
    11.111.111.11   myapp.default.11.112.113.14.xip.io
    ```