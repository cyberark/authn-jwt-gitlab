#!/usr/bin/env groovy

// Performs release promotion.  No other stages will be run
if (params.MODE == "PROMOTE") {
  release.promote(params.VERSION_TO_PROMOTE) { sourceVersion, targetVersion, assetDirectory ->
    // Any assets from sourceVersion Github release are available in assetDirectory
    // Any version number updates from sourceVersion to targetVersion occur here
    // Any publishing of targetVersion artifacts occur here
    // Anything added to assetDirectory will be attached to the Github Release
  }
  return
}

pipeline {
  agent { label 'executor-v2' }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '30'))
  }

  stages {
    // Generates a VERSION file based on the current build number and latest version in CHANGELOG.md
    // stage('Validate Changelog and set version') {
    //   steps {
    //     updateVersion("CHANGELOG.md", "${BUILD_NUMBER}")
    //   }
    // }

    stage('Get latest upstream dependencies') {
      steps {
        updateGoDependencies("${WORKSPACE}/go.mod")
      }
    }

    stage('Build while unit testit testing') {
      parallel {
        stage('Golang 1.19') {
          steps {
            sh './bin/test.sh'
          }
        }
      }
    }

  stage('Build release artifacts') {
    steps {
        sh "./bin/build_container_images"
      }
    }

    stage('Push images to internal registry') {
      steps {
        // Push images to the internal registry so that they can be used
        // by tests, even if the tests run on a different executor.
        sh './bin/publish-images internal'
      }
    }
    stage('Scan Docker Image') {
        parallel {
            stage("Scan Ubuntu Docker Image for fixable issues") {
                steps {
                    scanAndReport(containerImageWithTag_ubuntu(), "HIGH", false)
                      }
            }
            stage("Scan Ubuntu Docker image for total issues") {
                steps {
                    scanAndReport(containerImageWithTag_ubuntu(), "NONE", true)
                      }
            }

            stage("Scan UBI Docker Image for fixable issues") {
                steps {
                      scanAndReport(containerImageWithTag_ubi(), "HIGH", false)
                      }
            }
            stage("Scan UBI Docker image for total issues") {
                steps {
                    scanAndReport(containerImageWithTag_ubi(), "NONE", true)
                      }
            }

            stage("Scan Alpine Docker Image for fixable issues") {
                steps {
                    scanAndReport(containerImageWithTag_apline(), "HIGH", false)
                      }
            }
            stage("Scan Alpine Docker image for total issues") {
                steps {
                      scanAndReport(containerImageWithTag_apline(), "NONE", true)
                      }
            }
        }
    }

  }

  post {
    always {
      cleanupAndNotify(currentBuild.currentResult)
    }
  }
}


def containerImageWithTag_ubuntu() {
  sh(
    returnStdout: true,
    script: 'source ./bin/build_utils && echo "authn-jwt-gitlab:$(project_version_with_commit_alpine)"'
  )
}

def containerImageWithTag_ubi() {
  sh(
    returnStdout: true,
    script: 'source ./bin/build_utils && echo "authn-jwt-gitlab:$(project_version_with_commit_ubuntu)"'
  )
}

def containerImageWithTag_apline() {
  sh(
    returnStdout: true,
    script: 'source ./bin/build_utils && echo "authn-jwt-gitlab:$(project_version_with_commit_ubi)"'
  )
}

def containerImageWithTag() {
  var1 = $1
  sh 'echo Cyberark testing ${var1}'
  sh(
    returnStdout: true,
    script: 'source ./bin/build_utils && echo "authn-jwt-gitlab:${var1}$(project_version_with_commit)"'
  )
}

def tagWithSHA() {
  sh(
    returnStdout: true,
    script: 'echo $(git rev-parse --short=8 HEAD)'
  )
}

def versioning() {
  sh(
    returnStdout: true,
    script: 'echo 1.0.0'
  )
}