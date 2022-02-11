# Working on automation
**Today you will need access to our operations repo**
**Today you will need access to our azure stack**

## Build operations image
```sh
make docker-build-dev-platform-operations
```

## Create a customer
-  Pass in custom image, useful for development

```sh
docker run \
    --rm \
    -e OPERATIONS_IMAGE=platform-operations:latest \
    --entrypoint=/bin/sh \
    platform-operations:latest \
    -c '/app/bin/app tools jobs create-customer --customer-name="Test1"'
```

## Create a Application
-  Pass in custom image, useful for development

```sh
docker run \
    --rm \      
    -e OPERATIONS_IMAGE=platform-operations:latest \
    --entrypoint=/bin/sh \
    platform-operations:latest \
    -c '/app/bin/app tools jobs create-customer --customer-name="Test1"'
```