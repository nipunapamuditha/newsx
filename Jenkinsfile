pipeline {
    agent any
    
    tools {
        go 'go 1.23.4' // Use the Go version matching your go.mod
    }
    
    environment {
        GO111MODULE = 'on'
    }
    
    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }
        
        stage('Build') {
            steps {
                // Use 'go' directly, which will use the version defined in tools
                sh 'go build -o app'
                sh 'go test ./...'
            }
        }
        
        stage('Deploy') {
    steps {
        sshagent(['0ff14880-bcc6-4400-a835-a66a5a3cf0ba']) {
            sh '''
                ssh -o StrictHostKeyChecking=no jenkins@10.10.10.41 "echo SSH connection successful"
                ssh -o StrictHostKeyChecking=no jenkins@10.10.10.41 "mkdir -p /home/jenkins/newsx"
                scp -o StrictHostKeyChecking=no app jenkins@10.10.10.41:/home/jenkins/newsx
                scp -o StrictHostKeyChecking=no restart-service.sh jenkins@10.10.10.41:/home/jenkins/newsx
                ssh -o StrictHostKeyChecking=no jenkins@10.10.10.41 "cd /home/jenkins/newsx && chmod +x app restart-service.sh && ./restart-service.sh"
            '''
        }
    }
}
    }
    
    post {
        success {
            echo 'Deployment successful!'
        }
        failure {
            echo 'Build or deployment failed!'
        }
    }
}