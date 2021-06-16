# newkpt

## Description
sample description

## Usage

### Fetch the package
`kpt pkg get REPO_URI[.git]/PKG_PATH[@VERSION] newkpt`
Details: https://kpt.dev/reference/cli/pkg/get/

### View package content
`kpt pkg tree newkpt`
Details: https://kpt.dev/reference/cli/pkg/tree/

### Decrypt and render the package
```
kpt fn eval --image gcr.io/kpt-fn-contrib/sops:unstable --fn-config fnDecrypt.yaml --include-meta-resources --image-pull-policy ifNotPresent
kpt fn render
```

Note: `--image-pull-policy ifNotPresent` is needed only because we're using image from [this branch](https://github.com/aodinokov/kpt-functions-catalog/tree/allParams)

### Apply the package
```
kpt live init newkpt
kpt live apply newkpt --reconcile-timeout=2m --output=table
```
Details: https://kpt.dev/reference/cli/live/

### Encrypt and render the package before merging changes to git
```
kpt fn eval --image gcr.io/kpt-fn-contrib/sops:unstable --fn-config fnEncrypt.yaml --include-meta-resources --image-pull-policy ifNotPresent
kpt fn render
```
