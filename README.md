# Wordpress Demo

## Intro
- The purpose of the demo is to convey why Teleport is a good solution for securing access to Kubernetes clusters
- The demo will cover an infrastructure setup without Teleport
- The shortcommings of the infrastructure
- How Teleport overcomes the shortcommings

## Outline of infrastructure setup
- There are 3 virtual machines that consist of one master and two worker nodes
    - k8s-control
    - k8s-worker1
    - k8s-worker2 
- Ubuntu 20.04 is the operating system for all virtual machines
- The following is installed on each machine
    - cri-dockerd
    - Docker
    - kubeadm, kubctl and kubelet
- The following will be installed only on the master node
    - Helm v3
    - Calico CNI network plugin
- The Kubernetes infrastructure deployed on the worker nodes
    - Kubernetes Dashboard with a read only user
        - Readonly user is a service account, that has a secret of type token and cluster wide read only privileges
        - The Dashboard will be exposed via port forwarding
        - The token is pasted into the Kubernetes Dashboard for sign in     
    - Wordpress app
    - The Wordpress app will be deployed by a user with a kube-config file that only has access to the wordpress namespace
    - The Wordpress app will be deployed by a different user with a kube-config file that has port-forwarding access in the wordpress namespace to expose the wordpress Cluster ip service
    - Golang app with an API (Incomplete - Golang API has minimal functionality as explained in the installation notes)
        - The API will grab data for wordpress posts from the Wordpress MySQL database
        - The API image will be a locally built docker image, therefore the image pull policiy will be never for the pod running the API            

### cri-dockerd replaces dockershim


## Installation instructions

__Documentation for reference__
- Installing cri-dockerd: https://github.com/Mirantis/cri-dockerd/releases
- Dockershim and Cri-dockerd: https://www.mirantis.com/blog/cri-dockerd-faq-blog
- Installing Docker: https://docs.docker.com/engine/install/ubuntu/
- Installing kubeadm: https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/
- Creating a cluster with kubeadm: https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/
- Installing Helm v3: https://helm.sh/docs/intro/install/
- Deploy and Access Kubernetes Dashboard: https://kubernetes.io/docs/tasks/access-application-cluster/web-ui-dashboard/
- Deploying WordPress and MySQL with Persistent Volumes: https://kubernetes.io/docs/tutorials/stateful-application/mysql-wordpress-persistent-volume/


__Create three servers with the following settings:__

- Distribution: Ubuntu 20.04 Focal Fossa LTS


__Set appropriate hostname for each node__

E.g. k8s-control, k8s-worker1 and k8s-worker2


__On the control plane node:__

sudo hostnamectl set-hostname k8s-control


__On the first worker node:__
```
sudo hostnamectl set-hostname k8s-worker1
```

__On the second worker node__
```
sudo hostnamectl set-hostname k8s-worker2
```

__On all nodes, set up the hosts file to enable all the nodes to reach each other using these hostnames:__
```
sudo vi /etc/hosts
```

__On all nodes, add the following at the end of the file. The private IP address for each node is required:__

<control plane node private IP> k8s-control
<worker node 1 private IP> k8s-worker1
<worker node 2 private IP> k8s-worker2


__Log out of all three servers and log back in to see these changes take effect__



__On all nodes, set up Docker Engine and containerd. Load some kernel modules and modify some system settings as part of this
process:__
```
cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF

sudo modprobe overlay
sudo modprobe br_netfilter
```

__sysctl params required by setup; params persist across reboots:__
```
cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF
```

__Apply sysctl params without reboot:__
```
sudo sysctl --system
```

__Set up the Docker Engine repository:__

__Add Docker's official GPG key:__
```
sudo apt-get update
sudo apt-get install ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc
```

__Add the repository to Apt sources:__
```
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
  
sudo apt-get update

VERSION_STRING=5:23.0.1-1\~ubuntu.20.04\~focal

sudo apt-get install -y docker-ce=$VERSION_STRING docker-ce-cli=$VERSION_STRING containerd.io docker-buildx-plugin docker-compose-plugin

sudo systemctl enable docker
```

__Add the Ubuntu 'user' to the docker group:__
```
sudo usermod -aG docker $USER
```

__Log out and log back in so that the group membership is re-evaluated. Run the docker command below with sudo to verify docker is up and running__
```
docker run hello-world
```

__Build a local docker image for the golang app on every node__
```
cd ~
git clone https://github.com/abdul-ak413/wordpress-demo.git
cd ~/wordpress-demo/golang-api-docker/
docker image build -t wp-demo-golangapp:v1 .
docker image ls
```

__Enable containerd:__
```
sudo systemctl enable containerd
```

__On all nodes, setup cri-dockerd__
```
cd ~
wget https://github.com/Mirantis/cri-dockerd/releases/download/v0.3.15/cri-dockerd_0.3.15.3-0.debian-bullseye_amd64.deb
sudo dpkg -i cri-dockerd_0.3.15.3-0.debian-bullseye_amd64.deb
sudo systemctl enable cri-docker.service
sudo systemctl enable cri-docker.socket
```

__On all nodes, disable swap:__
```
sudo swapoff -a
#Permanently disable swap partition
sudo nano /etc/fstab
##/swap.img      none    swap    sw      0       0
```

__On all nodes, install kubeadm, kubelet, and kubectl:__
```
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.30/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg

echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.30/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list

sudo apt-get update && sudo apt-get install -y kubelet=1.30.0-1.1 kubeadm=1.30.0-1.1 kubectl=1.30.0-1.1

sudo apt-mark hold kubelet kubeadm kubectl

sudo systemctl enable kubelet
```

__On the control plane node only, initialize the cluster and set up kubectl access:__
```
sudo kubeadm config images pull --cri-socket unix://var/run/cri-dockerd.sock --kubernetes-version 1.30.0

sudo kubeadm init --pod-network-cidr 10.0.0.0/16 --kubernetes-version 1.30.0 --cri-socket unix://var/run/cri-dockerd.sock

mkdir -p $HOME/.kube

sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config

sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

__Verify the cluster is working:__
```
kubectl get nodes
```

__Install the Calico network add-on:__
```
kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.25.0/manifests/calico.yaml
```

__Get the join command (this command is also printed during kubeadm init):__
```
kubeadm token create --print-join-command 
```

__Copy the join command from the control plane node. Run it on each worker node as root (i.e., with sudo ):__
```
sudo kubeadm join ... --cri-socket unix://var/run/cri-dockerd.sock
```

__On the control plane node, verify all nodes in the cluster are ready. Note that it may take a few moments for all of the nodes to enter the READY state:__
```
kubectl get nodes
```

__On the control plane, install helm v3 from Binary releases__
1. Download your desired version
2. Unpack it (tar -zxvf helm-v3.15.4-linux-amd64.tar.gz)
3. Find the helm binary in the unpacked directory, and move it to its desired destination (mv linux-amd64/helm /usr/local/bin/helm)

```
wget https://get.helm.sh/helm-v3.15.4-linux-amd64.tar.gz
tar -zxvf helm-v3.15.4-linux-amd64.tar.gz
sudo mv linux-amd64/helm /usr/local/bin/helm
helm version
```

## Kubernetes Dashboard

### Deploy the Kubernetes Dashboard using Helm
```
# Add kubernetes-dashboard repository
helm repo add kubernetes-dashboard https://kubernetes.github.io/dashboard/
# Deploy a Helm Release named "kubernetes-dashboard" using the kubernetes-dashboard chart
helm upgrade --install kubernetes-dashboard kubernetes-dashboard/kubernetes-dashboard --create-namespace --namespace kubernetes-dashboard --set kong.admin.tls.enabled=false


#To access the dashboard, https://<contol plane private ip>:8443/
kubectl -n kubernetes-dashboard port-forward --address 0.0.0.0 svc/kubernetes-dashboard-kong-proxy 8443:443 &
```

### Create Read-Only Service Account for the Kubernetes Dashboard and retrieve the token to for login
```
cd ~/wordpress-demo/helm-charts
helm install readonly-1 read-only-dashboard/

#Retrive bearer token to sign into Kubernetes Dashboard
kubectl get secret secret-readonly -n kubernetes-dashboard -o jsonpath={".data.token"} | base64 -d
```


## Creating users in the Kubernetes cluster and granting access

### Create wordpress-dev user to deploy wordpress
```
mkdir ~/k8s-users/wordpress-dev -p

cd ~/k8s-users/wordpress-dev

openssl genrsa -out wordpress-dev.key 2048

openssl req -new -key wordpress-dev.key -subj "/CN=wordpress-dev" -out wordpress-dev.csr
```
```
cat <<EOF > csr_template.yaml
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: wordpress-dev.csr
spec:
  request: <Base64_encoded_CSR>
  signerName: kubernetes.io/kube-apiserver-client
  usages:
  - client auth
EOF
```

```
CSR_CONTENT=$(cat wordpress-dev.csr | base64 | tr -d '\n')

sed "s|<Base64_encoded_CSR>|$CSR_CONTENT|" csr_template.yaml > wordpress-dev_csr.yaml

kubectl create -f wordpress-dev_csr.yaml
```

```
kubectl get csr
kubectl certificate approve wordpress-dev.csr
kubectl get csr
kubectl get csr wordpress-dev.csr -o jsonpath='{.status.certificate}' | base64 --decode > wordpress-dev.crt
```

```
# Set Cluster Configuration:
kubectl config set-cluster kubernetes --server=https://192.168.1.110:6443 --certificate-authority=/etc/kubernetes/pki/ca.crt --embed-certs=true --kubeconfig=wordpress-dev.kubeconfig
```

```
kubectl config set-credentials wordpress-dev --client-certificate=wordpress-dev.crt --client-key=wordpress-dev.key --embed-certs=true --kubeconfig=wordpress-dev.kubeconfig
# Set wordpress-dev-context Context: 
kubectl config set-context wordpress-dev-context --cluster=kubernetes --namespace=wordpress --user=wordpress-dev --kubeconfig=wordpress-dev.kubeconfig
# Use wordpress-dev-context Context:
kubectl config use-context wordpress-dev-context --kubeconfig=wordpress-dev.kubeconfig
```

### Create wordpress-pf user to access wordpress via port forwarding

```
mkdir ~/k8s-users/wordpress-pf -p

cd ~/k8s-users/wordpress-pf

openssl genrsa -out wordpress-pf.key 2048

openssl req -new -key wordpress-pf.key -subj "/CN=wordpress-pf" -out wordpress-pf.csr
```
```
cat <<EOF > csr_template.yaml
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: wordpress-pf.csr
spec:
  request: <Base64_encoded_CSR>
  signerName: kubernetes.io/kube-apiserver-client
  usages:
  - client auth
EOF
```

```
CSR_CONTENT=$(cat wordpress-pf.csr | base64 | tr -d '\n')

sed "s|<Base64_encoded_CSR>|$CSR_CONTENT|" csr_template.yaml > wordpress-pf_csr.yaml

kubectl create -f wordpress-pf_csr.yaml
```

```
kubectl get csr
kubectl certificate approve wordpress-pf.csr
kubectl get csr
kubectl get csr wordpress-pf.csr -o jsonpath='{.status.certificate}' | base64 --decode > wordpress-pf.crt
```

```
# Set Cluster Configuration:
kubectl config set-cluster kubernetes --server=https://192.168.1.110:6443 --certificate-authority=/etc/kubernetes/pki/ca.crt --embed-certs=true --kubeconfig=wordpress-pf.kubeconfig
```

```
kubectl config set-credentials wordpress-pf --client-certificate=wordpress-pf.crt --client-key=wordpress-pf.key --embed-certs=true --kubeconfig=wordpress-pf.kubeconfig
# Set Developer Context: 
kubectl config set-context wordpress-pf-context --cluster=kubernetes --namespace=wordpress --user=wordpress-pf --kubeconfig=wordpress-pf.kubeconfig
# Use Developer Context:
kubectl config use-context wordpress-pf-context --kubeconfig=wordpress-pf.kubeconfig
```

### Grant access to the users wordpress-dev and wordpress-pf
__The helm charts contain cluster roles, roles and cluster role bindings and role bindings for the users__ 
```
cd ~/wordpress-demo/helm-charts
helm install rbac-1 rbac_wordpress_users/
```

### Deploy Wordpress as the wordpress-dev user
```
cd ~/wordpress-demo/helm-charts
helm install app-1 wordpress_app/ --kubeconfig=/home/$USER/k8s-users/wordpress-dev/wordpress-dev.kubeconfig
```

### Run the wordpress app deployed via port forwarding as the wordpress-pf user
```
kubectl port-forward --address 0.0.0.0 deployment/app-1-wordpress 8888:80 --kubeconfig=/home/dev-user/k8s-users/wordpress-pf/wordpress-pf.kubeconfig &

#Open the wordpress application on a webbrowser using the url http://<control-plane ip address>:8888
```

## Deploy Golang API (INCOMPLETE)
### Golang API connects to MySQL database server
```
#Limitations and errors of code
#Will throw an error and exit if wordpress.wp_posts table is not created
#wordpress.wp_posts table is only created after wordpress is installed
#API only retrieves wordpress posts at the time of creation. Post added after creation cannot be retrieved
#Custom wordpress posts are duplicated 

cd ~/wordpress-demo/helm-charts
helm install golang-api-1 golang-api/

#Retrieve data for wordpress posts including id
kubectl exec -n wordpress wp-golangapp exec -- curl 127.0.0.1:3000/posts
#Use wordpress post id to only retrieve data for a single wordpress post
kubectl exec -n wordpress wp-golangapp exec -- curl 127.0.0.1:3000/posts/post/<post id>
```

## Demo Infrastructure Shortcommings
- Overall a very large attack surface on a Kubernetes Cluster
- Secrets are exposed in a base64 format which can easily be decoded
- Require alot of effort to create a certficate for a user and approve the certficate for the Kubernetes cluster and distribute the certifcates
    - Lack of flexibilty: E.g. Users may require access to prod for an hour  
- Auditing will need to be setup

## Teleport 
- Single pane of glass view for the Kubernetes Cluster
- Very Flexible with Built in RBAC system
    - E.g. Temporary access to prod can easily be provided 
- Overall reduces the attack surface on a Kubernetes Cluster
- Users can be created in Teleport with certain priveleges, without the need for administrator to manage certificates
    - Teleport's Built-in RBAC system     
- Auditing in place
