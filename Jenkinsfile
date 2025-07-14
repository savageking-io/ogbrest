pipeline {
    agent {
        node {
            label 'Go Builder'
        }
    }

    stages {
        stage('Build') {
            steps {
                echo 'Building..'
                sh 'go get'
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
                echo 'Deploying....'
            }
        }
    }
}