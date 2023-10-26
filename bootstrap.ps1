docker build -t swervo_base -f ./Dependencies.Dockerfile .
docker build -t swervo_builder .
docker run --rm -v ${PWD}/build/bin:/artifacts swervo_builder