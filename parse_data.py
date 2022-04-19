from os import error
import sys
import json
import os.path

def read_json(file_path: str) -> dict:
  if os.path.exists(file_path):
    f = open(file_path, 'r')
    try:
      json_dict = json.load(f)
      f.close()
      return json_dict
    except Exception:
      return None
  else:
    return None

def get_contributors(path: str) -> str:
  file_path = f"{path}.contributors.json"
  contributors = read_json(file_path)
  if contributors is None:
    return ''

  return len(contributors)

def get_repodata(path: str) -> dict:
  dir_path = path.split('/')[1:11]
  dir_path = '/'.join(dir_path)
  file_path = f"/{dir_path}/repoMetadata.json"
  temp = read_json(file_path)
  repo_data = dict()
  if temp is None:
    repo_data['owner_type'] = ''
    repo_data['language'] = ''
    repo_data['size'] = ''
    return repo_data
  else:
    repo_data['owner_type'] = temp['owner']['type']
    repo_data['language'] = temp['language']
    repo_data['size'] = temp['size']
    return repo_data

def get_year(path: str) -> str:
  file_path = f"{path}.lastCommit.json"
  last_commit = read_json(file_path)
  if last_commit is None:
    return ""

  if len(last_commit) == 0:
    return ""
  else:
    return last_commit[0]['commit']['committer']['date'][:4]

def get_smells(path: str) -> list:
  file_path = f"{path}.hadolint.json"
  smell_list = read_json(file_path)
  if smell_list is None:
    return []

  smells = []
  for smell in smell_list:
    temp = dict()
    temp['code'] = smell['code']
    temp['level'] = smell['level']
    temp['type'] = smell['code'][:2]
    if temp['level'] in ['error', 'warning']:
      smells.append(temp)

  return smells

def get_filename(path: str) -> str:
  filename = path.split('/')[9:]
  return '/'.join(filename)

def get_name(path: str) -> str:
  author_repo = path.split('/')[9:11]
  return '/'.join(author_repo)

def get_loc(path: str) -> int:
  count = 0
  with open(path, encoding="latin1") as f:
    try:
      for line in f.readlines():
        line = line.strip()
        if line != '' and not line.startswith('#'):
          count += 1
    except error:
      pass
  return count

def get_size(path: str) -> int:
    st = os.stat(path)
    return st.st_size

def get_base_image(path: str) -> list:
  lines = []
  with open(path, encoding="latin1") as f:
    try:
      for line in f.readlines():
        line = line.strip()
        if line != '' and not line.startswith('#'):
          lines.append(line)
    except error:
      pass

  lines.reverse()
  for line in lines:
    if line.upper().startswith('FROM'):
      if ' ' in line:
        base_image = line.split(' ')[1]
      else:
        base_image = line
      return [base_image, base_image.split(':')[0]]
  return ['', '']

def parse_to_csv(paths: list[str]) -> list[str]:
  rows = []
  for path in paths:
    name = get_name(path)
    file_name = get_filename(path)
    file_size = get_size(path)
    loc = get_loc(path)
    smells = get_smells(path)
    year = get_year(path)
    repo_data = get_repodata(path)
    base_image = get_base_image(path)
    # TODO: add contributors to dockerfiles
    # contributors = get_contributors(path)
    for smell in smells:
      row = f"{name},{file_name},{smell['code']},{smell['type']},{smell['level']},{year},{file_size},{loc},{repo_data['owner_type']},{repo_data['language']},{repo_data['size']},{base_image[0]},{base_image[1]}"
      rows.append(row)
  return rows

def read_input() -> list[str]:
  return [line.strip() for line in sys.stdin]

def main():
  # reads paths to dockerfile dirs from stdin
  # in the following format: /path/to/dockerfile/dir/
  paths = read_input()
  rows = parse_to_csv(paths)
  # Different output formats
  # print('repo_name,year,file_size,loc,owner_type,language,size')
  # print('repo_name,file_name,smell_code,type,level,year,file_size,loc,owner_type,language,size')
  print('repo_name,file_name,smell_code,type,level,year,file_size,loc,owner_type,language,size,base_image_full,base_image')
  print('\n'.join(rows))

if __name__ == '__main__':
  main()