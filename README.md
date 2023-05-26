# authn-jwt-gitlab

## Description
This project creates a Docker image that includes a Go binary that can be used to authenticate a JWT token against Conjur Secrets Manager and retrieve a secret value.  Ubuntu, Alpine, and UBI-FIPS versions are available.  The secret value is returned to STDOUT and can be used in a GitLab CI pipeline.

## Badges
[![](https://img.shields.io/docker/pulls/nfmsjoeg/authn-jwt-gitlab)](https://hub.docker.com/r/nfmsjoeg/authn-jwt-gitlab) [![](https://img.shields.io/discord/802650809246154792)](https://discord.gg/J2Tcdg9tmk) [![](https://img.shields.io/reddit/subreddit-subscribers/cyberark?style=social)](https://reddit.com/r/cyberark) ![](https://img.shields.io/github/license/cyberark/authn-jwt-gitlab)

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

```yaml
# Sample Conjur policies
  * authn-jwt-gitlab/policy-gitlab-jwt/policy1.yml
  * authn-jwt-gitlab/policy-gitlab-jwt/policy2.yml
  * authn-jwt-gitlab/policy-gitlab-jwt/policy3.yml

# Sample Conjur variables' values
  * conjur variable values add conjur/authn-jwt/gitlab/token-app-property 'namespace_path'
  * conjur variable values add conjur/authn-jwt/gitlab/identity-path 'gitlab-apps'
  * conjur variable values add conjur/authn-jwt/gitlab/issuer 'https://gitlab.com'
  * conjur variable values add conjur/authn-jwt/gitlab/jwks-uri 'https://gitlab.com/-/jwks/’
```

## Usage

1. Choose your GitLab Runner Docker container image based on your desired OS.  The following images are available:
   * authn-jwt-gitlab:ubuntu-1.0.0
   * authn-jwt-gitlab:alpine-1.0.0
   * authn-jwt-gitlab:ubi-1.0.0
2. Once a GitLab Runner Docker container is decided upon, include it in your GitLab CI Pipeline file.  The following example is for the authn-jwt-gitlab:ubuntu-1.0.0 image:
```yaml
ubuntu:
  id_tokens:
    ID_TOKEN_1:
      aud: https://gitlab.com
  variables:
    CONJUR_AUTHN_JWT_TOKEN: $ID_TOKEN_1
  image: authn-jwt-gitlab:ubuntu-1.0.0
```
3. Be sure to properly tag the job in the GitLab CI Pipeline file with the proper tag to run the job on the GitLab Runner .
4. Variables must be set in the GitLab CI Pipeline file for the GitLab Runner Docker container to consume.  Those environment variables are:
    * `CONJUR_APPLIANCE_URL`
    * `CONJUR_ACCOUNT`
    * `CONJUR_AUTHN_JWT_SERVICE_ID`
    * `CONJUR_AUTHN_JWT_TOKEN`
    * `CONJUR_SECRET_ID`
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
  image: authn-jwt-gitlab:ubuntu-1.0.0
  script:
    - export TEST_USERNAME=$(CONJUR_SECRET_ID="Dev-Team-credential1" /authn-jwt-gitlab)
    - export TEST_PASSWORD=$(CONJUR_SECRET_ID="Dev-Team-credential2" /authn-jwt-gitlab)
    - env | grep TEST_

alpine:
  id_tokens:
    ID_TOKEN_2:
      aud: https://gitlab.com
  variables:
    CONJUR_AUTHN_JWT_TOKEN: $ID_TOKEN_2
  image: authn-jwt-gitlab:alpine-1.0.0
  script:
    - export TEST_USERNAME=$(CONJUR_SECRET_ID="Dev-Team-credential1" /authn-jwt-gitlab)
    - export TEST_PASSWORD=$(CONJUR_SECRET_ID="Dev-Team-credential2" /authn-jwt-gitlab)
    - env | grep TEST_

ubi-fips:
  id_tokens:
    ID_TOKEN_3:
      aud: https://gitlab.com
  variables:
    CONJUR_AUTHN_JWT_TOKEN: $ID_TOKEN_3
  image: authn-jwt-gitlab:ubi-1.0.0
  script:
    - export TEST_USERNAME=$(CONJUR_SECRET_ID="Dev-Team-credential1" /authn-jwt-gitlab)
    - export TEST_PASSWORD=$(CONJUR_SECRET_ID="Dev-Team-credential2" /authn-jwt-gitlab)
    - env | grep TEST_
```

## Support
This is a community supported project.  For support, please file an issue in this repository.

## Contributing
If you would like to contribute to this project, please review the [CONTRIBUTING.md](CONTRIBUTING.md) file.

## License
This project is licensed under MIT - see the [LICENSE](LICENSE) file for details.
