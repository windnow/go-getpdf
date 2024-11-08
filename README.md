# Пример формирования PDF

## WSL: Установка Chrome

```sh
sudo sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google-chrome.list'
sudo apt update
sudo apt install google-chrome-stable
```

## Сборка Docker:

```sh
docker build . --tag getpdf
docker run -p 8080:8080 -d --name getpdf getpdf
```