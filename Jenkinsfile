#!/usr/bin/env groovy

// This is a template Jenkinsfile for builds and the automated release project

// Automated release, promotion and dependencies
properties([
  // Include the automated release parameters for the build
  release.addParams(),
  // Dependencies of the project that should trigger builds
  dependencies([])
])

// Performs release promotion.  No other stages will be run
if (params.MODE == "PROMOTE") {
  release.promote(params.VERSION_TO_PROMOTE) { sourceVersion, targetVersion, assetDirectory ->
    // Any assets from sourceVersion Github release are available in assetDirectory
    // Any version number updates from sourceVersion to targetVersion occur here
    // Any publishing of targetVersion artifacts occur here
    // Anything added to assetDirectory will be attached to the Github Release
    sh """docker pull registry.tld/authn-jwt-gitlab:ubuntu-${sourceVersion}
          docker pull registry.tld/authn-jwt-gitlab:alpine-${sourceVersion}
          docker pull registry.tld/authn-jwt-gitlab:ubi-${sourceVersion}
          docker tag registry.tld/authn-jwt-gitlab:ubuntu-${sourceVersion} authn-jwt-gitlab:ubuntu-${targetVersion}
          docker tag registry.tld/authn-jwt-gitlab:alpine-${sourceVersion} authn-jwt-gitlab:alpine-${targetVersion}
          docker tag registry.tld/authn-jwt-gitlab:ubi-${sourceVersion} authn-jwt-gitlab:ubi-${targetVersion}
       """
    sh "./publish-images --promote --version=${targetVersion}"
  }
  return
}

pipeline {
  agent { label 'executor-v2' }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '30'))
  }

  triggers {
    cron(getDailyCronString())
  }

  environment {
    // Sets the MODE to the specified or autocalculated value as appropriate
    MODE = release.canonicalizeMode()
  }

  stages {
    // Aborts any builds triggered by another project that wouldn't include any changes
    stage ("Skip build if triggering job didn't create a release") {
      when {
        expression {
          MODE == "SKIP"
        }
      }
      steps {
        script {
          currentBuild.result = 'ABORTED'
          error("Aborting build because this build was triggered from upstream, but no release was built")
        }
      }
    }
    // Generates a VERSION file based on the current build number and latest version in CHANGELOG.md
    stage('Validate Changelog and set version') {
      steps {
        updateVersion("CHANGELOG.md", "${BUILD_NUMBER}")
      }
    }

    stage('Get latest upstream dependencies') {
      steps {
        updateGoDependencies("${WORKSPACE}/go.mod")
      }
    }

    stage('Unit Tests') {
      environment {
        GO_VERSION = "1.19"
      }
      steps {
        sh './bin/test.sh'
      }
    }

    stage('Build Images') {
      steps {
        sh "./bin/build_container_images"
      }
    }
  
    stage('Scan Images') {
      environment {
        TAG = sh(returnStdout: true, script: "./bin/version_with_commit.sh")
      }
      parallel {
        stage("Scan Ubuntu Docker Image for fixable issues") {
          steps {
            scanAndReport("authn-jwt-gitlab:ubuntu-${env.TAG}", "HIGH", false)
          }
        }
        stage("Scan Ubuntu Docker image for total issues") {
          steps {
            scanAndReport("authn-jwt-gitlab:ubuntu-${env.TAG}", "NONE", true)
          }
        }
        stage("Scan UBI Docker Image for fixable issues") {
          steps {
            scanAndReport("authn-jwt-gitlab:ubi-${env.TAG}", "HIGH", false)
          }
        }
        stage("Scan UBI Docker image for total issues") {
          steps {
            scanAndReport("authn-jwt-gitlab:ubi-${env.TAG}", "NONE", true)
          }
        }
        stage("Scan Alpine Docker Image for fixable issues") {
          steps {
            scanAndReport("authn-jwt-gitlab:alpine-${env.TAG}", "HIGH", false)
          }
        }
        stage("Scan Alpine Docker image for total issues") {
          steps {
            scanAndReport("authn-jwt-gitlab:alpine-${env.TAG}", "NONE", true)
          }
        }
      }
    }

    // Push images to internal registry with associated commit hash
    stage('Push images to internal registry') {
      steps {
        sh './bin/publish-images --internal'
      }
    }

    stage('Release') {
      when {
        expression {
          MODE == "RELEASE"
        }
      }

      steps {
        release { billOfMaterialsDirectory, assetDirectory ->
          // Publish release artifacts to all the appropriate locations
          // Copy any artifacts to assetDirectory to attach them to the Github release
          sh './bin/publish-images --edge'
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