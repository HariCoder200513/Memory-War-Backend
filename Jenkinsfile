pipeline {
    agent any

    parameters {
        string(name: 'EC2_HOST', defaultValue: '', description: 'EC2 public IP or hostname to deploy to')
    }

    environment {
        BINARY_NAME = 'memory-wars-server'
        DEPLOY_DIR  = '/opt/memory-wars'
        EC2_USER    = 'ubuntu'
    }

    stages {

        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Setup') {
            steps {
                sh '''
                    go version
                    go mod download
                '''
            }
        }

        stage('Vet') {
            steps {
                sh 'go vet ./...'
            }
        }

        stage('Test') {
            steps {
                sh 'go test ./... -v -cover'
            }
        }

        stage('Build') {
            steps {
                sh '''
                    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
                        go build -ldflags="-s -w" \
                        -o ${BINARY_NAME} ./cmd/server
                '''
                archiveArtifacts artifacts: "${BINARY_NAME}", fingerprint: true
            }
        }

        stage('Deploy to EC2') {
            when {
                allOf {
                    branch 'main'
                    expression { params.EC2_HOST?.trim() }
                }
            }
            steps {
                withCredentials([
                    sshUserPrivateKey(credentialsId: 'mw-ec2-ssh-key', keyFileVariable: 'SSH_KEY'),
                    file(credentialsId: 'mw-backend-env', variable: 'BACKEND_ENV')
                ]) {
                    sh '''
                        chmod 600 "$SSH_KEY"
                        EC2="${EC2_USER}@${EC2_HOST}"

                        scp -o StrictHostKeyChecking=no -i "$SSH_KEY" \
                            ${BINARY_NAME} "${EC2}:${DEPLOY_DIR}/${BINARY_NAME}.new"

                        scp -o StrictHostKeyChecking=no -i "$SSH_KEY" \
                            "$BACKEND_ENV" "${EC2}:${DEPLOY_DIR}/.env.new"

                        ssh -o StrictHostKeyChecking=no -i "$SSH_KEY" "${EC2}" '
                            cd /opt/memory-wars
                            mv .env.new .env && chmod 600 .env
                            mv memory-wars-server.new memory-wars-server && chmod +x memory-wars-server
                            sudo systemctl restart memory-wars
                        '
                    '''
                }
            }
        }
    }

    post {
        success {
            echo 'Backend pipeline succeeded.'
        }
        failure {
            echo 'Backend pipeline failed.'
        }
        always {
            cleanWs()
        }
    }
}
