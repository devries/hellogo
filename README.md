# README #

This is a demonstration Docker container with go. Try it on Google Cloud by clicking below:

[![Open in Cloud
Shell](http://gstatic.com/cloudssh/images/open-btn.png)](https://console.cloud.google.com/cloudshell/open?git_repo=https://github.com/devries/hellogo.git&page=editor&open_in_editor=README.md)

### Testing the Container ###

1. Build the image using docker:
    ```
    docker build -t hellogo .
    ```
2. Run the container on port 8080:
    ```
    docker run --rm -p 8080:8080 hellogo
    ```
3. In the cloud shell right hand upper corner click "Web Preview > Preview
    on port 8080."

### Move the Container to Container Registry ###

1. You must have a project, and use a project name in the tags and calls
    below. We will use the sample name `my-project-name` in the examples below,
    but you should change this to match your project name.
2. Tag the image for google:
    ```
    docker tag hellogo gcr.io/my-project-name/hellogo:latest
    ```
3. Push to google container registry:
    ```
    gcloud docker -- push gcr.io/my-project-name/hellogo
    ```

