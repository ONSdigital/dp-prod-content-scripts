# dp-prod-content-scripts

Scripts for manipulating collections & site content.

# Warning :warning:  
Making direct changes to the live website content is risky and it's very easy to do a *"boo boo"* that  creates, updates
 or delete the wrong thing. Unless there is a specific reason not to it's **highly highly recommended** that your any 
 changes your script makes are done via a collection. __You have been warned__

## Running on an environment

Assuming you have written a script and tested on your local copy of the website - to run a script in a dev/production 
environment: _ssh_ on to the desired environment (see dp-setup). Run a _Golang_ docker container with a volume mapped 
to the content/master/whatever dir:

If a go container already exists start and connect:
```
sudo docker start -i <container_name>
```

Otherwise create a new container:

```
sudo docker run -i -t --name <container_name> \
   -v <path-to-content>:<path-in-container> \
   golang /bin/bash
``` 

| Var                 | Description                                                 |
|---------------------|-------------------------------------------------------------|
| _container_name_    | The name of the container                                   |
| _path-to-content_   | The directory path to the content dir on the environment    |
| _path-in-container_ | From the container this is how we reference the content dir |


Once the container is running install some useful tools:

```
// Get up to date
apt-get update

// editing files
apt-get install vim

// human viewable JSON
apt-get install jq
```

Get the code 
```
go get github.com/ONSdigital/dp-prod-content-scripts/countX
```
Move to the package dir you want to run.
```
cd go/src/github.com/ONSdigital/dp-prod-content-scripts/countX
```

execute your script
```
./run.sh
```

#### Accessing files on the mapped volume
When we created the container we defined a volume mount which maps a directory on the host env to a dir in the docker 
container. If we want to access a file in the mounted volume we do so using a path relative to the volume path we 
specified.

##### Example 

If _path-to-content_ is `/home/website/content` and _path-in-container_ is `/site` 
from within the container/our go script we access `/home/website/content/home.html` using `/site/home.html`