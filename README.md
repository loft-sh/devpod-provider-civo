# CIVO Provider for DevPod

## Getting started

The provider is available for auto-installation using

```sh
devpod provider add civo
devpod provider use civo
```

Follow the on-screen instructions to complete the setup.

Needed variables will be:

- CIVO_REGION
- CIVO_API_KEY

### Creating your first devpod env with civo

After the initial setup, just use:

```sh
devpod up .
```

You'll need to wait for the machine and environment setup.

### Customize the VM Instance

This provides has the seguent options

|    NAME            | REQUIRED |          DESCRIPTION                  |         DEFAULT         |
|--------------------|----------|---------------------------------------|-------------------------|
| CIVO_DISK_IMAGE    | false    | The disk image to use.                | d927ad2f-5073-4ed6-b2eb-b8e61aef29a8   |
| CIVO_DISK_SIZE     | false    | The disk size to use.                 | 40                       |
| CIVO_INSTANCE_TYPE | false    | The machine type to use.              | g3.xsmall                |
| CIVO_REGION        | true     | The civo cloud region to create the VM |                         |
| CIVO_API_KEY       | true     | The api key to use                    |                         |

Options can either be set in `env` or using for example:

```sh
devpod provider set-options -o CIVO_REGION=LON1
```
