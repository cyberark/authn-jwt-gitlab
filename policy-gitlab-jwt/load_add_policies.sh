#!/bin/bash

docker-compose exec -T client conjur policy load root policy/policy1.yml
docker-compose exec -T client conjur policy load root policy/policy2.yml
docker-compose exec -T client conjur policy load root policy/policy3.yml

docker-compose exec -T client conjur variable values add conjur/authn-jwt/gitlab/token-app-property 'namespace_path'
docker-compose exec -T client conjur variable values add conjur/authn-jwt/gitlab/identity-path 'gitlab-apps'
docker-compose exec -T client conjur variable values add conjur/authn-jwt/gitlab/issuer 'https://gitlab.com'
docker-compose exec -T client conjur variable values add conjur/authn-jwt/gitlab/jwks-uri 'https://gitlab.com/-/jwks/'

docker exec -it conjur_client conjur variable values add Dev-Team-credential1 "UserName_secret1"
docker exec -it conjur_client conjur variable values add Dev-Team-credential2 "Password_secret2"
docker exec -it conjur_client conjur variable values add Dev-Team-credential3 "UserName_secret3"
docker exec -it conjur_client conjur variable values add Dev-Team-credential4 "Password_secret4"