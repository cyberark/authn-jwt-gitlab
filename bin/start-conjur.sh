#!/bin/bash -ex

. ./utils
. ./build_utils

trap teardown ERR

main() {

  retrieve_cyberark_ca_cert
  announce "Pulling images..."
  docker-compose pull "conjur" "postgres" "cli5" "conjur-server"
  echo "Done!"

  announce "Building images..."
  docker-compose build "conjur" "postgres" "conjur-server"
  echo "Done!"

  announce "Starting Conjur environment..."
  export CONJUR_DATA_KEY="$(docker-compose run -T --no-deps conjur data-key generate)"
  docker-compose up --no-deps -d "conjur" "postgres" "conjur-server"
  echo "Done!"

  announce "Waiting for conjur to start..."
  exec_on conjur conjurctl wait

  echo "Done!"

  api_key=$(exec_on conjur conjurctl role retrieve-key cucumber:user:admin | tr -d '\r')

  # Export values needed for tests to access Conjur instance
  export CONJUR_AUTHN_API_KEY="$api_key"
  export CONJUR_APPLIANCE_URL="http://conjur"
  export CONJUR_ACCOUNT="cucumber"
  export CONJUR_AUTHN_LOGIN="admin"

}

main
