pipeline {
    agent {
        node {
            label 'Go Builder'
        }
    }

    stages {
        stage('Preparing') {
            steps {
                sh 'go mod download'
            }
        }
        stage('Build') {
            steps {
                echo 'Building..'
                sh 'make'
            }
        }
        stage('Test') {
            steps {
                echo 'Testing..'
                sh 'make test'
            }
        }
        stage('Deploy') {
            steps {
                when {
                    branch 'main'
                }
                steps {
                    script {
                        build job: 'OnlineGameBase/Docker/ogbrest-dev', wait: true
                    }
                }
            }
        }
    }
}