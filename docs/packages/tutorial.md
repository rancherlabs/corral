# How to Create a Package

For this tutorial we will create a Corral package that sets up an OCI registry with authentication on a Digitalocean 
droplet.

# Anatomy of a Package

To create a corral package we need to understand what goes into a package.  Packages consist of 3 components. The manifest
defines what information is displayed to end users and what the user is required to provide to use the package.  The 
terraform modules define the infrastructure required for this package. Finally, the overlay defines scripts and assets 
used to configure the applications running on the package infrastructure.

To start lets define the basic folder structure of our package.  Going forward we will refer to the registry folder as
the root of our package.

```shell
mkdir -p registry/{terraform/main,overlay}
```

# Defining the Manifest

The manifest tells corral how to create a corral from this package as well as how users should interact with it.  To start
lets give our package a name and description.  Create `manifest.yaml` in our package's root directory.  Our description
should describe what the package does as well as give the user a sense of what the package will create and how it should
be used.  Packages should be specific in their usage to make them as reproducible as possible.  It is better to create
many similar packages than a single customizable package.

```yaml
name: registry
description: >
  An authenticated docker registry running in docker for local development.
```

Now we can define how users will interact with this package.  We can do this with variables. variables define what values
can be passed into a package as well as what values should be displayed to a user.  While corral scripts and terraform
modules can define any variable they like, only the variables defined in the manifest will be exposed to the user. Lets
update our `manifest.yaml` to look like this.

```yaml
name: registry
description: |
  An authenticated docker registry running in Digitalocean.
variables:
  digitalocean_token:
    sensitive: true
    type: string
    optional: false
    description: "A Digitalocean API token with write permission. https://docs.digitalocean.com/reference/api/create-personal-access-token/"
  digitalocean_domain:
    sensitive: true
    type: string
    optional: false
    description: "The domain to use for the registry host."
  registry_host:
    type: string
    readOnly: true
    description: "host the configured registry can be accessed at"
  username:
    type: string
    readOnly: true
    description: "username for registry authentication"
  password:
    type: string
    readOnly: true
    description: "password for registry authentication"
```

As we can see the package requires Digitalocean credentials and a domain.  While these values may be different uesr to
user the will not fundamentally change the behavior of the package.  Any variables that change behavior should be
a different package.  We also define `registry_host`, `username`, and `password`  These values will be defined by Corral
when the package is used and cannot be changed by the user.

# Writing a Terraform Module

Now that we have defined how we want our package to work lets define the infrastructure we need to make it happen.  To start
lets create a file with all the variables available to us.  Any Corral variable will be exposed as a terraform variable
prefixed with `corral_`.

Create the file `terraform/main/corral.tf`.  It is best practice to define corral variables in `corral.tf`
```terraform
// Corral
variable "corral_name" {} // name of the corral being created
variable "corral_user_id" {} // how the user is identified (usually github username)
variable "corral_user_public_key" {} // the users public key
variable "corral_public_key" {} // The corrals public key.  This should be installed on every node.
variable "corral_private_key" {} // The corrals private key.

// Package
variable "digitalocean_token" {}
variable "digitalocean_domain" {}
```

Now that we have defined what values are available we can create our registry's infrastructure.

`terraform/main/main.tf`

```terraform
terraform {
  required_version = ">= 0.13"
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

provider "random" {}
provider "digitalocean" {
  token = var.digitalocean_token
}

// it is best practice to distinguish an environment with a random id to avoid collisions
resource "random_id" "registry_id" {
  byte_length       = 6
}

// we will use the corral public key to get access to nodes to provision them later
resource "digitalocean_ssh_key" "corral_key" {
  name       = "${var.corral_user_id}-${random_id.registry_id.hex}"
  public_key = var.corral_public_key
}

resource "digitalocean_droplet" "registry" {
  count = 1

  name = "${var.corral_user_id}-${random_id.registry_id.hex}-registry"
  image    = "ubuntu-20-04-x64"
  region   = "sfo3"
  size     = "s-1vcpu-2gb"
  tags = [var.corral_user_id, random_id.registry_id.hex] // when possible resources should be marked with the associated corral
  ssh_keys = [digitalocean_ssh_key.corral_key.id]
}

resource "digitalocean_record" "registry" {
  domain = var.digitalocean_domain
  name   = random_id.registry_id.hex
  type   = "A"
  value  = digitalocean_droplet.registry[0].ipv4_address
}
```

Now that we have all this infrastructure we need to tell corral how to interact with our infrastructure.  We need
to define our node pools.  Node pools are grouping of ssh hosts that corral can execute commands on. In addition to
the node pools we can set the registry host here as we have everything we need to define it.  Any terraform output  will
be stored as a corral variable.

Let's create `terraform/main/outputs.tf`

```terraform
output "corral_node_pools" {
  value = {
    registry = [
    for droplet in digitalocean_droplet.registry : {
      name = droplet.name // unique name of node
      user = "root" // ssh username
      address = droplet.ipv4_address // address of ssh host
    }
    ]
  }
}

output "registry_host" {
  value = join(".", [digitalocean_record.registry.name, digitalocean_record.registry.domain])
}
```

# Overlay

Now that we have some infrastructure to work with we can configure our application.  By default, the overlay directory
will be copied to the root directory of all nodes.  All files will be copied with the ownership of the ssh user in mode
`0777`.  Best practice is to put any scripts used only for provisioning the nodes in `/opt/corral`.  For the purposes
of this tutorial we can just copy the overlay directory from `examples/registry/overlay` in this repository.  This
contains the registry binary and some other assets need for the application.  Most of these files do not interact with 
corral but `overlay/opt/corral/install.sh` takes advantage of some Corral shell features.

```shell
#!/bin/bash
set -ex

# corral_set allows us to set corral variables from scripts.
function corral_set() {
    echo "corral_set $1=$2"
}

# corral_log allows us to print messages for the corral user.
function corral_log() {
    echo "corral_log $1"
}

# Install the user's public key incase they need to debug an issue
echo "$CORRAL_corral_user_public_key" >> /$(whoami)/.ssh/authorized_keys

apt install -y apache2-utils


USERNAME="corral"
PASSWORD="$( echo $RANDOM | md5sum | head -c 12)"  # it is best practice to generate passwords for every distinct corral

# here we set the username and password for the user to find later
corral_set username $USERNAME
corral_set password $PASSWORD

# this will be used by the docker registry for authentication
htpasswd -Bbn $USERNAME "$PASSWORD" > /etc/docker/registry/htpasswd

# corral variables are available as environment variables with the prefix `CORRAL_`
sed -i "s/HOSTNAME/$CORRAL_registry_host/g" /etc/docker/registry/config.yml

# generate self signed certificates
openssl req -x509 \
            -newkey rsa:4096 \
            -sha256 \
            -days 3650 \
            -nodes \
            -keyout /etc/docker/registry/ssl/registry.key \
            -out /etc/docker/registry/ssl/registry.crt \
            -subj "/CN=${CORRAL_registry_host}" \
            -addext "subjectAltName=DNS:${CORRAL_registry_host}"

corral_log "This registry uses self signed certificates please add {\"insecure_registries\":[\"${CORRAL_registry_host}\"]} to /etc/docker/daemon.json."

systemctl enable registry
systemctl start registry
```

# Commands

The last step is to tell corral to run our terraform module in the manifest.  We do this with the commands section.
Commands can either be a terraform module or a shell command to run on a node pool.  If there are multiple nodes in a 
node pool the commands will be run concurrently. To run our terraform module let's add a command section to our 
`manifest.yaml`

```yaml
commands:
  - module: main
```

Now that our terraform is run we can run commands against our registry node to configure the registry. You can have as
commands as you want but for this package we only need to run a single install script.

```yaml
commands:
  - module: main
  - command: /opt/corral/install.sh
    node_pools:
      - registry
```

# Validating a Package

At this point we have configured our manifest, infrastructure and scripts to configure our application.  We should now
have a valid corral package that we can create corrals from!  Before we try to create a corral we can validate our 
package.

```shell
corral package validate ./registry
```

If we have any typos in our manifest or the folder structure has any problems this command will output them.


# Installing a Local Package

Assuming our package validated we can now test it!  If there are any issues corral will automatically rollback the
infrastructure.  This is convenient for keeping cloud environments clean but can make it difficult to diagnose issues.
We can tell corral to pipe stdout and stdin to our terminal while we create our package to better understand any issues
we may encounter with the `--debug` flag.  We can also pass our Digitalocean credentials and domain with the `-v` flag.
For variables like these it is better to set them as global variables, so we don't need to type them out every time we
create a corral from a package that uses Digitalocean.

```shell
corral config vars set digitalocean_token MY_DO_TOKEN
corral config vars set digitalocean_domain my.domain.example.com
```

```shell
corral create registry ./registry --debug
```

Once our package is created we can see that our registry host and credentials are ready to use.
```shell
corral vars registry
```
