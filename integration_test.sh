docker-compose -f deployments/docker-compose.tests.yaml up --build
EXIT_CODE=$?
exit ${EXIT_CODE}