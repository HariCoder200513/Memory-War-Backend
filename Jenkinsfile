pipeline {
    agent any

    parameters {
        string(name: 'EC2_HOST', defaultValue: '', description: 'EC2 instance public IP or hostname to deploy to')
    }

    environment {
        GO_VERSION     = '1.23'
        APP_NAME       = 'memory-wars-backend'
        BINARY_NAME    = 'memory-wars-server'
        DEPLOY_DIR     = '/opt/memory-wars'
        EC2_USER       = 'ubuntu'
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Setup Go') {
            steps {
                sh '''
                    go version
                    go mod download
                '''
            }
        }

        stage('Lint & Vet') {
            steps {
                sh '''
                    go vet ./...
                '''
            }
        }

        stage('Test') {
            steps {
                withCredentials([
                    string(credentialsId: 'mw-database-url', variable: 'DATABASE_URL'),
                    string(credentialsId: 'mw-jwt-secret', variable: 'JWT_SECRET')
                ]) {
                    sh '''
                        go test ./... -v -cover
                    '''
                }
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
                    sshUserPrivateKey(credentialsId: 'mw-ec2-ssh-key', keyFileVariable: 'SSH_KEY', usernameVariable: 'SSH_USER'),
                    string(credentialsId: 'mw-database-url', variable: 'DATABASE_URL'),
                    string(credentialsId: 'mw-jwt-secret', variable: 'JWT_SECRET'),
                    string(credentialsId: 'mw-google-client-id', variable: 'GOOGLE_CLIENT_ID'),
                    string(credentialsId: 'mw-smtp-host', variable: 'SMTP_HOST'),
                    string(credentialsId: 'mw-smtp-port', variable: 'SMTP_PORT'),
                    string(credentialsId: 'mw-smtp-user', variable: 'SMTP_USER'),
                    string(credentialsId: 'mw-smtp-password', variable: 'SMTP_PASSWORD'),
                    string(credentialsId: 'mw-smtp-from', variable: 'SMTP_FROM'),
                    string(credentialsId: 'mw-frontend-base-url', variable: 'FRONTEND_BASE_URL')
                ]) {
                    sh '''
                        chmod 600 "$SSH_KEY"

                        EC2_TARGET="${EC2_USER}@${EC2_HOST}"

                        # Copy binary to EC2
                        scp -o StrictHostKeyChecking=no -i "$SSH_KEY" \
                            ${BINARY_NAME} ${EC2_TARGET}:${DEPLOY_DIR}/${BINARY_NAME}.new

                        # Generate .env on EC2 and restart
                        ssh -o StrictHostKeyChecking=no -i "$SSH_KEY" ${EC2_TARGET} << 'DEPLOY_SCRIPT'
                            set -e

                            cd /opt/memory-wars

                            # Write .env from Jenkins credentials
                            cat > .env << ENVEOF
PORT=8080
DATABASE_URL=${DATABASE_URL}
JWT_SECRET=${JWT_SECRET}
JWT_EXPIRY=15m
REFRESH_EXPIRY=168h
GOOGLE_CLIENT_ID=${GOOGLE_CLIENT_ID}
SMTP_HOST=${SMTP_HOST}
SMTP_PORT=${SMTP_PORT}
SMTP_USER=${SMTP_USER}
SMTP_PASSWORD=${SMTP_PASSWORD}
SMTP_FROM=${SMTP_FROM}
FRONTEND_BASE_URL=${FRONTEND_BASE_URL}
ENVIRONMENT=production
ENVEOF
                            chmod 600 .env

                            # Swap binary atomically
                            mv memory-wars-server.new memory-wars-server
                            chmod +x memory-wars-server

                            # Restart systemd service
                            sudo systemctl restart memory-wars
DEPLOY_SCRIPT
                    '''
                }
            }
        }
    }

    post {
        success {
            echo "Backend build and deploy succeeded."
        }
        failure {
            echo "Backend pipeline failed."
        }
        always {
            cleanWs()
        }
    }
}
