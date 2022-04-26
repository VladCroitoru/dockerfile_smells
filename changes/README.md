# Changes

Contains input data and results used to argue about the evolution of Dockerfiles.

Some quick useful commands:

```sh
# count the number of changes.json files
$ find changes/gha -type f -name '*.json' | wc -l
$ find changes/sep -type f -name '*.json' | wc -l

# list changes.json files and the number of commits it contains
$ find changes/gha -type f -name '*.json' | ./count_commits.sh
$ find changes/sep -type f -name '*.json' | ./count_commits.sh

# fetch commits and save the result to a changes/{gha,sep}/$author/$repo/changes.json file
$ ./fetch_changes.sh changes/gha < changes/input_gha.txt
$ ./fetch_changes.sh changes/sep < changes/input_sep.txt
```
