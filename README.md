# Conjur Gitlab Plugin

## Description
This project creates a Docker image that includes a Go binary that can be used to authenticate a JWT against Conjur Secrets Manager and retrieve a secret value.  Ubuntu, Alpine, and UBI-FIPS versions are available.  
## Certification level
[![](https://img.shields.io/badge/Certification%20Level-Certified-28A745?)](https://github.com/cyberark/community/blob/master/Conjur/conventions/certification-levels.md)

This repository is a **Certified** level project. It's a community contributed project **reviewed and tested by CyberArk
and trusted to use with Conjur Open Source**. For more detailed information on our certification levels, see [our community guidelines](https://github.com/cyberark/community/blob/master/Conjur/conventions/certification-levels.md#certified).

## Requirements

* Docker and access to DockerHub
* Golang 1.20.2+
* GitLab on Docker (gitlab/gitlab-ce:15.11.2-ce.0+)
* GitLab Runner on Docker (gitlab/gitlab-runner:v15.8.2+)
* Conjur OSS 1.9+
* Conjur Enterprise 12.5+

## Setup Self hosted GitLab on docker
Provide GITLAB_ADDRESS value without http/https and port number , for example gitlab.example.com

```yaml
#!/bin/bash
#============ Variables ===============
# Is sudo required to run docker/podman - leave empty if no need
SUDO=
# Using docker/podman
CONTAINER_MGR=docker
# Docker image URL
CONTAINER_IMG=gitlab/gitlab-ce:latest
# GitLab URL (if available, use the external hostname)
GITLAB_ADDRESS=
# GitLab HTTP port
GITLAB_HTTP_PORT=9080
# GitLab admin user password
GITLAB_ROOT_PASSWORD=pqr@123
#============ Deploying GitLab ===============
$SUDO $CONTAINER_MGR run --detach \
  --hostname "$GITLAB_ADDRESS" \
  --publish "$GITLAB_HTTP_PORT":$GITLAB_HTTP_PORT \
  --name gitlab-server \
  --restart always \
  --shm-size 1gb \
  --env GITLAB_ROOT_PASSWORD="$GITLAB_ROOT_PASSWORD" \
  --env GITLAB_OMNIBUS_CONFIG="external_url 'http://$GITLAB_ADDRESS:$GITLAB_HTTP_PORT/';" \
  "$CONTAINER_IMG"
```
## Setup Self hosted GitLab Runner on docker
To get GITLAB_REGISTRATION_TOKEN value , create project in Gitlab and Go to your project -> settings -> CI/CD -> Runner. From Runner section get GITLAB_REGISTRATION_TOKEN and GITLAB_HOST details .

```yaml
#!/bin/bash
#============ Variables ===============
# Is sudo required to run docker/podman - leave empty if no need
SUDO=
# Using docker/podman
CONTAINER_MGR=docker
# Docker image URL
CONTAINER_IMG=gitlab/gitlab-runner:latest
# GitLab Host
GITLAB_HOST=$(hostname -f)
# GitLab port
GITLAB_PORT=9080
# GitLab Instance Registration Token
GITLAB_REGISTRATION_TOKEN=
#============ Deploying GitLab Runner ===============
$SUDO $CONTAINER_MGR volume create gitlab-runner-conjur-config
$SUDO $CONTAINER_MGR run -d --name gitlab-runner-conjur --restart always \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v gitlab-runner-conjur-config:/etc/gitlab-runner \
    "$CONTAINER_IMG"
# Registering GitLab runner
$SUDO $CONTAINER_MGR run --rm -it -v gitlab-runner-conjur-config:/etc/gitlab-runner \
    "$CONTAINER_IMG" register -u "http://$GITLAB_HOST:$GITLAB_PORT" -r "$GITLAB_REGISTRATION_TOKEN" \
    --description "Demo Runner" -n --tag-list "conjur-demo" --executor docker --docker-image alpine:latest
# Deploying Summon on the runner
$SUDO $CONTAINER_MGR exec -it gitlab-runner-conjur bash -c 'curl -sSL https://raw.githubusercontent.com/cyberark/summon/main/install.sh | bash'
$SUDO $CONTAINER_MGR exec -it gitlab-runner-conjur bash -c 'curl -sSL https://raw.githubusercontent.com/cyberark/summon-conjur/main/install.sh | bash'

```
## Keep Conjur Server up and running with policies settings

### Conjur Setup

   JWT Authenticator is required at Conjur server.  You may wish to refer to [official doc](https://docs.cyberark.com/Product-Doc/OnlineHelp/AAM-DAP/Latest/en/Content/Operations/Services/cjr-authn-jwt-uc.htm?tocpath=Integrations%7CJWT%20Authenticator%7C_____1)

### Conjur policies setup

* Create a Conjur policy for the JWT Authenticator
  
  - Save the following policy as authn-jwt.yml:
  ```yaml
  - !policy
    id: conjur/authn-jwt/gitlab
    annotations:
      description: JWT Authenticator web service for gitlab
      gitlab: true
    body:
      # Create the conjur/authn-jwt/gitlab web service
      - !webservice

      # Optional: Uncomment any or all of the following variables:
      # * token-app-propery
      # * identity-path  conjur/authn-jwt/gitlab/identity-path  gitlab/root
      # * issuer
      # identity-path is always used together with token-app-property
      # however, token-app-property can be used without identity-path

      - !variable
        id: token-app-property
        annotations:
          description: JWT Authenticator bases authentication on claims from the JWT. You can base authentication on identifying clams such as the name, the user, and so on. If you can customize the JWT, you can create a custom claim and base authentication on this claim.

      - !variable
        id: identity-path
        annotations:
          description: JWT Authenticator bases authentication on a combination of the claim in the token-app-property and the full path of the application identity (host) in Conjur. This variable is optional, and is used in conjunction with token-app-property. It has no purpose when standing alone.

      - !variable
        id: issuer
        annotations:
          description: JWT Authenticator bases authentication on the JWT issuer. This variable is optional, and is relevant only if there is an iss claim in the JWT. The issuer variable and iss claim values must match.

      - !variable
        id: jwks-uri

        ## Group of hosts that can authenticate using this JWT Authenticator
      - !group
        id: apps

      # Permit the consumers group to authenticate to the authn-jwt/gitlab web service
      - !permit
        role: !group apps
        privilege: [ read, authenticate ]
        resource: !webservice

      # Create a web service for checking authn-jwt/gitlab status
      - !webservice
        id: status

      # Group of users who can check the status of authn-jwt/gitlab
      - !group
        id: operators

      # Permit group to check the status of authn-jwt/gitlab
      - !permit
        role: !group operators
        privilege: read
        resource: !webservice status
  ```
    
    - Load the policy into root:
  ```
    conjur policy load -f /path/to/file/authn-jwt.yml -b root
  ```   
  
* Populate the policy variables

```yaml
  * conjur variable set -i conjur/authn-jwt/gitlab/token-app-property -v 'namespace_path'
  * conjur variable set -i conjur/authn-jwt/gitlab/identity-path -v 'gitlab-apps'
  * conjur variable set -i conjur/authn-jwt/gitlab/issuer -v 'https://gitlab.com'
  * conjur variable set -i conjur/authn-jwt/gitlab/jwks-uri -v 'https://gitlab.com/-/jwks/’
```

* Define an app ID (host)
```yaml
- !policy
  id: gitlab-apps
  body:

      # Group of hosts that can authenticate using this JWT Authenticator
    - !group

        # `gitlab_name` is the primary identifying claim
    - &hosts
      - !host
        id: myapp
        annotations:
          description: Host identity for authn-jwt-gitlab project in root namespace within GitLab
          authn-jwt/gitlab/ref: main
          authn-jwt/gitlab/project_path: myapp/authn-jwt-gitlab

    # Grant all hosts in collection above to be members of projects group
    - !grant
      role: !group
      members: *hosts

- !grant
  role: !group conjur/authn-jwt/gitlab/apps
  member: !group gitlab-apps
```

    - Load the policy into root:
  ```
    conjur policy load -f /path/to/file/authn-jwt-hosts.yml -b root
  ```  

* Secret Variables and Permissions:

```yaml
- &devvariables
   - !variable Dev-Team-credential1
   - !variable Dev-Team-credential2
   - !variable Dev-Team-credential3
   - !variable Dev-Team-credential4

- !permit
  resource: *devvariables
  privileges: [ read, execute ]
  roles: !group gitlab-apps
```
  - Load the policy into root:
  ```
    conjur policy load -f /path/to/file/authn-jwt-secret-variables.yml -b root
  ``` 

* Set the secret variable   
     a. Generate a secret

     Generate a value for your application’s secret:
     ```
     secretVal=$(openssl rand -hex 12 | tr -d '\r\n')
     ```

     This generates a 12-hex-character value.

     b. Store the secret

     Store the generated value in Conjur:
     ```
     conjur variable set -i Dev-Team-credential1 -v ${secretVal}
     ```

* Allowlist the JWT Authenticator
<small><a href='https://docs.cyberark.com/Product-Doc/OnlineHelp/AAM-DAP/Latest/en/Content/Operations/Services/authentication-types.htm#Configur'>For details, see Configure authenticators. </a></small>
  - Environment variable
```
CONJUR_AUTHENTICATORS=authn-jwt/gitlab
```

## Usage

1. Choose your GitLab Runner Docker container image based on your desired OS.  The following images are available:
   * cyberark/authn-jwt-gitlab:ubuntu-1.0.0
   * cyberark/authn-jwt-gitlab:alpine-1.0.0
   * cyberark/authn-jwt-gitlab:ubi-1.0.0
2. Once a GitLab Runner Docker container is decided upon, include it in your GitLab CI Pipeline file.  The following example is for the cyberark/authn-jwt-gitlab:ubuntu-1.0.0 image:
```yaml
ubuntu:
  id_tokens:
    ID_TOKEN_1:
      aud: https://gitlab.com
  variables:
    CONJUR_AUTHN_JWT_TOKEN: $ID_TOKEN_1
  image: cyberark/authn-jwt-gitlab:ubuntu-1.0.0
```
3. Be sure to properly tag the job in the GitLab CI Pipeline file with the proper tag to run the job on the GitLab Runner .
4. Variables must be set in the GitLab CI Pipeline file for the GitLab Runner Docker container to consume.  Those environment variables are:
    * `CONJUR_APPLIANCE_URL`
    * `CONJUR_ACCOUNT`
    * `CONJUR_AUTHN_JWT_SERVICE_ID`
    * `CONJUR_AUTHN_JWT_TOKEN`
    * `CONJUR_SECRET_ID`
    * `CONJUR_SSL_CERTIFICATE` or `CONJUR_CERT_FILE`
5. To use the binary in a job executing on the GitLab Runner , review the [example GitLab CI Pipeline script](.gitlab-ci.yml) in this repository.

### Example GitLab CI YAML File

```yaml
variables:
  CONJUR_APPLIANCE_URL: "https://conjur_server"
  CONJUR_ACCOUNT: "myConjurAccount"
  CONJUR_AUTHN_JWT_SERVICE_ID: "gitlab"

ubuntu:
  id_tokens:
    ID_TOKEN_1:
      aud: https://gitlab.com
  variables:
    CONJUR_AUTHN_JWT_TOKEN: $ID_TOKEN_1
  image: cyberark/authn-jwt-gitlab:ubuntu-1.0.0
  script:
    - export TEST_USERNAME=$(CONJUR_SECRET_ID="Dev-Team-credential1" /authn-jwt-gitlab)
    - export TEST_PASSWORD=$(CONJUR_SECRET_ID="Dev-Team-credential2" /authn-jwt-gitlab)

alpine:
  id_tokens:
    ID_TOKEN_2:
      aud: https://gitlab.com
  variables:
    CONJUR_AUTHN_JWT_TOKEN: $ID_TOKEN_2
  image: cyberark/authn-jwt-gitlab:alpine-1.0.0
  script:
    - export TEST_USERNAME=$(CONJUR_SECRET_ID="Dev-Team-credential1" /authn-jwt-gitlab)
    - export TEST_PASSWORD=$(CONJUR_SECRET_ID="Dev-Team-credential2" /authn-jwt-gitlab)

ubi-fips:
  id_tokens:
    ID_TOKEN_3:
      aud: https://gitlab.com
  variables:
    CONJUR_AUTHN_JWT_TOKEN: $ID_TOKEN_3
  image: cyberark/authn-jwt-gitlab:ubi-1.0.0
  script:
    - export TEST_USERNAME=$(CONJUR_SECRET_ID="Dev-Team-credential1" /authn-jwt-gitlab)
    - export TEST_PASSWORD=$(CONJUR_SECRET_ID="Dev-Team-credential2" /authn-jwt-gitlab)
```

## Support
This is a community supported project.  For support, please file an issue in this repository.

## Contributing
If you would like to contribute to this project, please review the [CONTRIBUTING.md](CONTRIBUTING.md) file.

## License
This project is licensed under MIT - see the [LICENSE](LICENSE) file for details.
