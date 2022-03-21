
import sys
import requests
from bs4 import BeautifulSoup as Soup

def get_html_page(repo: str, page: int) -> list[str, str]:
  r = requests.get(f'https://github.com/{repo}/search?l=dockerfile&p={page}')
  soup = Soup(r.text, 'html.parser')
  header = Soup.select(soup, '#repo-content-pjax-container > div > div > div.col-12.col-md-9.float-left.px-2.pt-3.pt-md-0.codesearch-results > div > div.d-flex.flex-column.flex-md-row.flex-justify-between.border-bottom.pb-3.position-relative > h3')
  return [soup, header]

def get_page_list(page_count: int) -> list[int]:
  pages = []
  for _ in range(page_count):
    pages.append(10)

  last_page = page_count % 10
  if last_page != 0:
    pages.append(last_page)

  return pages

def parse_repos(repos: list[str]) -> list[str]:
  for repo in repos:
    try:
      dockerfiles = parse_repo(repo)
      print('\n'.join(dockerfiles))
    except Exception:
      print(repo, file=sys.stderr)

def parse_repo(repo: str) -> str:
  [_, header] = get_html_page(repo, 1)
  count = int(header[0].text.rstrip().split(" ")[4])

  page_count = int(count / 10)
  pages = get_page_list(page_count)
  
  dockerfiles = []
  for idx, page in enumerate(pages):
    [html, _] = get_html_page(repo, idx)
    for i in range(1, page+1):
      selector = f"#code_search_results > div > div:nth-child({i}) > div > div.f4.text-normal > a"
      anchor = Soup.select(html, selector)
      file_path = anchor[0].text
      dockerfile = f"{count},{repo},{file_path}"
      dockerfiles.append(dockerfile)

  return dockerfiles

def read_input() -> list[str]:
  return [line.strip() for line in sys.stdin]

def main():
  repos = read_input()
  parse_repos(repos)

if __name__ == '__main__':
  main()
