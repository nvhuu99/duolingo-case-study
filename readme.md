### Start:
    * Docker engine must be running

    Start minikube:         
        minikube start --driver=docker --nodes=1 --memory=6g --cpus=4

### Build images

    Load images:  
        kind load docker-image message-input:latest --name kind-default-cluster

    Build images:                           
        docker build --build-arg SVC_DIR_NAME="message_input" -t message-input -f ./docker/dockerfile ./

### Deploy:

    Start all deployments:          
        kubectl apply -f ./k8s/

    Verify:                         
        kubectl config set-context --current --namespace=duolingo-case-study-namespace
        
        kubectl get all

        kubectl logs -l app=message-input

### Operations:

    RabbitMQ Dashboard (Available at: localhost:15672):         
        kubectl port-forward svc/rabbitmq 15672:15672
                                
    MongoDB Compass:            
        kubectl port-forward svc/mongodb-svc 27017:27017

    Message Input Api:
        kubectl port-forward svc/message-input 80:80

        kubectl logs -f --tail=-1 -l app=message-input

### Termination:

    kubectl delete all --all -n duolingo-case-study-namespace

### Others:

    kubectl rollout restart deployment <deployment-name>