# Fileinfo

This application solves a common problem in HTTP API backends - reading metadata from user-uploaded content.

There are several issues with implementing a solution which reads a file into memory within your own application. It is advantageous to abstract this from your business logic by letting a service such as S3 handle file upload and storage. You can use this application as a microservice to read metadata by passing it a URL to the file. With S3 this would be a signed URL.

It is build using my [runtime](https://github.com/g-wilson/runtime) library for easy deployment on Lambda behind API Gateway.

It uses FFProbe and Exiftool as dependencies for some "analyzer" types. On Lambda they should be deployed as Lambda Layers.
