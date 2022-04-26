# Composition

Contains the composition of datasets. The composition is used to argue about the similarity or difference of the datasets prior to comparing the results of the static quality assessment by the `hadolint` tool.

- repo_name - repository name.
- year - year of the last commit to the tracked Dockerfile.
- file_size - size of the Dockerfile.
- loc - lines of code of the Dockerfile, comments are ignored.
- owner_type - owner type of the repository, can be either user or organization.
- language - main language of the repository.
- size - size of the repository.
- base_image_full - base image of the Dockerfile, full identifier.
- base_image - base image of the Dockerfile, version/tag removed.
