# async-component
Async component for knative services

![diagram](./README-images/diagram.png)

## Install Knative Serving & Eventing to your Cluster
  1. https://knative.dev/docs/install/any-kubernetes-cluster/

<!--- 
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
      
-->

## Create your Demo Application. 

  1. This can be any simple hello world application that sleeps for some time. There is a sample application that writes to a cloudant DB in the `test/app` folder. To deploy, update the credentials in the `service.yml` and deploy the application with `kubectl deploy -f service.yml`.

  1. Make note of your application URL.

## Install the Consumer and Producer
  1. Apply the following config files:

      ```
      ko apply -f config/async-requests/100-async-consumer.yaml
      ko apply -f config/async-requests/100-async-producer.yaml
      ```

## Install the Redis Source

  1. Follow the `Getting Started` Instructions for the [Redis Source](https://github.com/lionelvillard/eventing-redis)

  1. For the `Example` section, do not install the entire `samples` folder, as you don't need the event-display sink. Additionally, edit the `redisstream.yaml` file to point to your consumer as the sink. The change should look something like this:
     
    ```
    apiVersion: sources.knative.dev/v1alpha1
    kind: RedisStreamSource
    metadata:
      name: mystream
    spec:
      address: "redis.redis.svc.cluster.local:6379"
      stream: mystream
      sink:
        ref:
          apiVersion: v1
          kind: Service
          name: async-consumer
    ```

  1. Apply the changed file to your cluster:
    ```
    kubectl apply -f samples/redisstream.yaml
    ```
  
## Test the Application
  1. Update your vs.yaml file with the URL to your application under `hosts`. 

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