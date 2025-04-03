pipeline {
    agent any
    
    tools {
        go 'go' // Ensure Go is configured in Jenkins Global Tool Configuration
    }
    
    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }
        
        stage('Build') {
            steps {
                sh 'go build -o app'
                sh 'go test ./...'
            }
        }
        
        stage('Deploy') {
            steps {
                sshagent(['0ff14880-bcc6-4400-a835-a66a5a3cf0ba']) { // Replace with your Jenkins SSH credentials ID
                    // Copy the built application to your node
                    sh '''
                        ssh jenkins@10.10.10.41 "mkdir -p /home/jenkins/newsx"
                        scp app jenkins@10.10.10.41:/home/jenkins/newsx
                        ssh jenkins@10.10.10.41 "cd /home/jenkins/newsx && chmod +x app && ./restart-service.sh"
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