# GO Coding Task

We have a small GO coding task for you, to complete before meeting,  
which we will then talk about during meeting:

## Task

Create an API to accept a link to a helm chart, search for  
container images. Download docker images from their respective  
repositories and return a response of the list of images, their size  
and no. of layers in each image.

## Hints for developers

- Helm docs https://helm.sh/docs/  
- k8s docs https://kubernetes.io/docs/concepts/containers/  
- Helm chart example https://github.com/helm/examples  
- Running Helm template

```bash
git clone https://github.com/helm/examples.git  
cd examples/charts/hello-world  
helm template . -f values.yaml
