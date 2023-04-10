# authn-jwt-gitlab

## Description
This project creates a Docker image that includes a Go binary that can be used to authenticate a JWT token against Conjur Secrets Manager and retrieve a secret value.  Ubuntu, Alpine, and UBI-FIPS versions are available.  The secret value is returned to STDOUT and can be used in a GitLab CI pipeline.

## Badges
[![](https://img.shields.io/docker/pulls/nfmsjoeg/authn-jwt-gitlab)](https://hub.docker.com/r/nfmsjoeg/authn-jwt-gitlab) [![](https://img.shields.io/discord/802650809246154792)](https://discord.gg/J2Tcdg9tmk) [![](https://img.shields.io/reddit/subreddit-subscribers/cyberark?style=social)](https://reddit.com/r/cyberark) ![](https://img.shields.io/github/license/infamousjoeg/authn-jwt-gitlab)

## Requirements

* [Docker GitLab Runner](https://docs.gitlab.com/runner/install/docker.html)
* [Conjur Secrets Manager](https://www.conjur.org)
* Conjur Policies for authentication & authorization (authn & authz)
  * [authn-jwt Conjur Policy with GitLab Service ID](https://github.com/infamousjoeg/conjur-policies/tree/master/authn/authn-jwt-gitlab.yml)
  * [Conjur Policy to create identity for GitLab Repository](https://github.com/infamousjoeg/conjur-policies/blob/16f7375b604646a48b8b59ac9ddc011b6c8a08c6/ci/gitlab/root.yml#L45)
  * [Conjur Policy to grant GitLab Repository identity to use synchronized secrets from CyberArk Vault](https://github.com/infamousjoeg/conjur-policies/blob/84b451b5025fd1bb5fc86c601d172cb27da81b00/grants/grants_ci.yml#L41)
  * [Conjur Policy to grant GitLab Repository identity ability to authenticate using authn-jwt/gitlab web service](https://github.com/infamousjoeg/conjur-policies/blob/84b451b5025fd1bb5fc86c601d172cb27da81b00/grants/grants_authn.yml#L23)
* [Setup GitLab with runner](#setup-gitlab-with-runner)
* [Keep Conjur Server up and running with policies settings](#keep-conjur-server-up-and-running-with-policies-settings)
* [Usage](#usage)

## Setup GitLab with runner
1. Start your [Free GitLab Ultimate trial](https://about.gitlab.com/free-trial/) for 30 days
2. Add project in GitLab
3. Get GitLab runner registration token and url details
4. Make sure settings -> CI/CD -> Token Access is disabled , runner is active and enable runner to pick jobs without tags .
5. Setup runner
```yaml
docker run -d --name gitlab-runner-temp --restart always \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /etc/gitlab-runner \
  gitlab/gitlab-runner:latest
```
6. Register runner
```yaml
docker exec -it gitlab-runner-temp gitlab-runner register
```
  * To take runner registration token and url , go to specificproject -> settings -> CI/CD -> Runners
  * Executor - shell , Tag - ConjurDemo , Description - DemoRunner

## Keep Conjur Server up and running with policies settings
### Setup Conjur server
* [Conjur setup](https://github.com/cyberark/conjur-quickstart)
* Conjur policies setup
```yaml
  * authn-jwt-gitlab/policy-gitlab-jwt/policy1.yml
  * authn-jwt-gitlab/policy-gitlab-jwt/policy2.yml
  * authn-jwt-gitlab/policy-gitlab-jwt/policy3.yml
  * authn-jwt-gitlab/policy-gitlab-jwt/load_add_policies.sh
```
## Usage

1. Choose your GitLab Runner Docker container image based on your desired OS.  The following images are available:
   * nfmsjoeg/authn-jwt-gitlab:ubuntu
   * nfmsjoeg/authn-jwt-gitlab:alpine
   * nfmsjoeg/authn-jwt-gitlab:ubi-fips
2. Once a GitLab Runner Docker container is decided upon, include it in your GitLab CI Pipeline file.  The following example is for the nfmsjoeg/authn-jwt-gitlab:ubuntu image:
```yaml
ubuntu:
  id_tokens:
    ID_TOKEN_1:
      aud: https://gitlab.com
  variables:
    CONJUR_AUTHN_JWT_TOKEN: $ID_TOKEN_1
  image: nfmsjoeg/authn-jwt-gitlab:ubuntu
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
  CONJUR_APPLIANCE_URL: "http://35.223.135.74:8080"
  CONJUR_ACCOUNT: "myConjurAccount"
  CONJUR_AUTHN_JWT_SERVICE_ID: "gitlab"

ubuntu:
  id_tokens:
    ID_TOKEN_1:
      aud: https://gitlab.com
  variables:
    CONJUR_AUTHN_JWT_TOKEN: $ID_TOKEN_1
  image: nfmsjoeg/authn-jwt-gitlab:ubuntu
  script:
    - export TEST_USERNAME=$(CONJUR_SECRET_ID="Dev-Team-credential1" /authn-jwt-gitlab)
    - export TEST_PASSWORD=$(CONJUR_SECRET_ID="Dev-Team-credential2" /authn-jwt-gitlab)
    - echo $ID_TOKEN_1 | base64
    - env | grep TEST_

alpine:
  id_tokens:
    ID_TOKEN_2:
      aud: https://gitlab.com
  variables:
    CONJUR_AUTHN_JWT_TOKEN: $ID_TOKEN_2
  image: nfmsjoeg/authn-jwt-gitlab:alpine
  script:
    - export TEST_USERNAME=$(CONJUR_SECRET_ID="Dev-Team-credential1" /authn-jwt-gitlab)
    - export TEST_PASSWORD=$(CONJUR_SECRET_ID="Dev-Team-credential2" /authn-jwt-gitlab)
    - echo $ID_TOKEN_2 | base64
    - env | grep TEST_

ubi-fips:
  id_tokens:
    ID_TOKEN_3:
      aud: https://gitlab.com
  variables:
    CONJUR_AUTHN_JWT_TOKEN: $ID_TOKEN_3
  image: nfmsjoeg/authn-jwt-gitlab:ubi-fips
  script:
    - export TEST_USERNAME=$(CONJUR_SECRET_ID="Dev-Team-credential1" /authn-jwt-gitlab)
    - export TEST_PASSWORD=$(CONJUR_SECRET_ID="Dev-Team-credential2" /authn-jwt-gitlab)
    - echo $ID_TOKEN_3 | base64
    - env | grep TEST_
```

## Support
This is a community supported project.  For support, please file an issue in this repository.

## Contributing
If you would like to contribute to this project, please review the [CONTRIBUTING.md](CONTRIBUTING.md) file.

## License
This project is licensed under MIT - see the [LICENSE](LICENSE) file for details.
