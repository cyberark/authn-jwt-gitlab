#!/usr/bin/env groovy

@Library("product-pipelines-shared-library") _

def productName = 'Conjur Authentication JWT Gitlab'
def productTypeName = 'Conjur Enterprise'

properties([
  // Include the automated release parameters for the build
  release.addParams(),
  // Dependencies of the project that should trigger builds
  dependencies([])
])

if (params.MODE == "PROMOTE") {
  release.promote(params.VERSION_TO_PROMOTE) { infrapool, sourceVersion, targetVersion, assetDirectory ->

    env.INFRAPOOL_PRODUCT_NAME = "${productName}"
    env.INFRAPOOL_DD_PRODUCT_TYPE_NAME = "${productTypeName}"

    // Scan the images before promoting
    def scans = [:]
    
    scans["Scan Ubuntu Docker image"] = {
      runSecurityScans(infrapool,
        image: "registry.tld/authn-jwt-gitlab:ubuntu-${sourceVersion}",
        buildMode: params.MODE,
        branch: env.BRANCH_NAME,
        arch: 'linux/amd64')
    }

    scans["Scan Alpine Docker image"] = {
      runSecurityScans(infrapool,
            image: "registry.tld/authn-jwt-gitlab:alpine-${sourceVersion}",
            buildMode: params.MODE,
            branch: env.BRANCH_NAME,
            arch: 'linux/amd64')
    }

    scans["Scan Ubi Docker image"] = {
      runSecurityScans(infrapool,
            image: "registry.tld/authn-jwt-gitlab:ubi-${sourceVersion}",
            buildMode: params.MODE,
            branch: env.BRANCH_NAME,
            arch: 'linux/amd64')
    }

    parallel scans

    // Pull existing images from internal registry in order to promote
    infrapool.agentSh """
      export PATH="release-tools/bin:${PATH}"
      docker pull registry.tld/authn-jwt-gitlab:ubuntu-${sourceVersion}
      docker pull registry.tld/authn-jwt-gitlab:alpine-${sourceVersion}
      docker pull registry.tld/authn-jwt-gitlab:ubi-${sourceVersion}
      # Promote source version to target version.
      ./bin/publish-images promote --source-version ${sourceVersion} --target-version ${targetVersion}
    """

    // Ensure the working directory is a safe git directory for the subsequent
    // promotion operations after this block.
    sh 'git config --global --add safe.directory "$(pwd)"'
  }
  // Copy Github Enterprise release to Github
  release.copyEnterpriseRelease(params.VERSION_TO_PROMOTE)
  return
}

pipeline {
  agent { label 'conjur-enterprise-common-agent' }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '30'))
  }
  environment {
    // Sets the MODE to the specified or autocalculated value as appropriate
    MODE = release.canonicalizeMode()
  }

  triggers {
    cron(getDailyCronString())
    parameterizedCron(getWeeklyCronString("H(1-5)","%MODE=RELEASE"))
  }
  stages {
    stage('Get latest upstream dependencies') {
      steps {
        sh './bin/updateGoDependencies.sh -g "${WORKSPACE}/go.mod"'
      }
    }
    stage('Get InfraPool ExecutorV2 Agent') {
      steps {
        script {
          INFRAPOOL_EXECUTORV2_AGENT_0 = getInfraPoolAgent.connected(type: "ExecutorV2", quantity: 1, duration: 1)[0]
        }
      }
   }
   
    stage('Build while unit testing') {
      parallel {
        stage('Golang 1.23.7') {
          steps {
            script {
              INFRAPOOL_EXECUTORV2_AGENT_0.agentSh './bin/test.sh'
              INFRAPOOL_EXECUTORV2_AGENT_0.agentStash name: 'output-xml', includes: 'output/1.23.7/*.xml'
            }
          }
        }
      }
    }
    // Generates a VERSION file based on the current build number and latest version in CHANGELOG.md
    stage('Validate changelog and set version') {
      steps {
        updateVersion(INFRAPOOL_EXECUTORV2_AGENT_0, "CHANGELOG.md", "${BUILD_NUMBER}")
      }
    }
    stage('Build release artifacts') {
      steps {
        script {
          INFRAPOOL_EXECUTORV2_AGENT_0.agentSh "./bin/build_container_images"        
        }
      }
    }

    stage('Scan Docker Image') {
      parallel {
        stage("Scan Ubuntu Docker Image for fixable issues") {
          steps {
            scanAndReport(INFRAPOOL_EXECUTORV2_AGENT_0, containerImageWithTag_ubuntu(), "HIGH", false)
          }
        }
        stage("Scan Ubuntu Docker image for total issues") {
          steps {
            scanAndReport(INFRAPOOL_EXECUTORV2_AGENT_0, containerImageWithTag_ubuntu(), "NONE", true)
          }
        }
        stage("Scan UBI Docker Image for fixable issues") {
          steps {
            scanAndReport(INFRAPOOL_EXECUTORV2_AGENT_0, containerImageWithTag_ubi(), "HIGH", false)
          }
        }
        stage("Scan UBI Docker image for total issues") {
          steps {
            scanAndReport(INFRAPOOL_EXECUTORV2_AGENT_0, containerImageWithTag_ubi(), "NONE", true)
          }
        }
        stage("Scan Alpine Docker Image for fixable issues") {
          steps {
            scanAndReport(INFRAPOOL_EXECUTORV2_AGENT_0, containerImageWithTag_alpine(), "HIGH", false)
          }
        }
        stage("Scan Alpine Docker image for total issues") {
          steps {
            scanAndReport(INFRAPOOL_EXECUTORV2_AGENT_0, containerImageWithTag_alpine(), "NONE", true)
          }
        }
      }
    }

    stage('Push images to internal registry') {
      steps {
        script {
          INFRAPOOL_EXECUTORV2_AGENT_0.agentSh './bin/publish-images internal'
        }
      }
    }

    stage('Release') {
      when {
        allOf {
          branch 'main' 
          expression { MODE == "RELEASE" }
        }
      }
      steps {
        script {
          release(INFRAPOOL_EXECUTORV2_AGENT_0) { billOfMaterialsDirectory, assetDirectory, toolsDirectory ->
            INFRAPOOL_EXECUTORV2_AGENT_0.agentSh """
              export PATH="${toolsDirectory}/bin:${PATH}" &&
              go-bom --tools "${toolsDirectory}" --go-mod ./go.mod --image "golang" --output "${billOfMaterialsDirectory}/go-mod-bom.json"
            """
            INFRAPOOL_EXECUTORV2_AGENT_0.agentSh "./bin/publish-images edge"
          }
        }
      }
    }
  }

  post {
    always {
      unstash 'output-xml'
      junit 'output/1.23.7/junit.xml'
      cobertura autoUpdateHealth: false, autoUpdateStability: false,
                coberturaReportFile: 'output/1.23.7/coverage.xml',
                conditionalCoverageTargets: '30, 0, 0',
                lineCoverageTargets: '30, 0, 0',
                methodCoverageTargets: '30, 0, 0',
                failUnhealthy: false, failUnstable: false,
                onlyStable: false, sourceEncoding: 'ASCII',
                zoomCoverageChart: false
      codacy action: 'reportCoverage', filePath: "output/1.23.7/coverage.xml"

      releaseInfraPoolAgent(".infrapool/release_agents")
    }
  }
}

def containerImageWithTag_ubuntu() {
  INFRAPOOL_EXECUTORV2_AGENT_0.agentSh(
    returnStdout: true,
    script: 'source ./bin/build_utils && echo "authn-jwt-gitlab:$(project_version_with_commit_ubuntu)"'
  )
}

def containerImageWithTag_ubi() {
  INFRAPOOL_EXECUTORV2_AGENT_0.agentSh(
    returnStdout: true,
    script: 'source ./bin/build_utils && echo "authn-jwt-gitlab:$(project_version_with_commit_ubi)"'
  )
}

def containerImageWithTag_alpine() {
  INFRAPOOL_EXECUTORV2_AGENT_0.agentSh(
    returnStdout: true,
    script: 'source ./bin/build_utils && echo "authn-jwt-gitlab:$(project_version_with_commit_alpine)"'
  )
}

def containerImageWithTag() {
  var1 = $1
  sh 'echo Cyberark testing ${var1}'
  INFRAPOOL_EXECUTORV2_AGENT_0.agentSh(
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