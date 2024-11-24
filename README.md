# OpenFGA Operator

[![Release](https://github.com/ZEISS/openfga-operator/actions/workflows/release.yml/badge.svg)](https://github.com/ZEISS/openfga-operator/actions/workflows/release.yml)
[![Taylor Swift](https://img.shields.io/badge/secured%20by-taylor%20swift-brightgreen.svg)](https://twitter.com/SwiftOnSecurity)
[![Volkswagen](https://auchenberg.github.io/volkswagen/volkswargen_ci.svg?v=1)](https://github.com/auchenberg/volkswagen)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

OpenFGA Operator is a Kubernetes operator for managing OpenFGA deployments.

## Installation

[Helm](https://helm.sh/) can be used to install the `openfga-operator` to your Kubernetes cluster.

```shell
helm repo add openfga-operator https://zeiss.github.io/openfga-operator/helm/charts
helm repo update
helm search repo openfga-operator
```

## License

[Apache 2.0](/LICENSE)