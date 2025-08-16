import json
import requests
import os

API_URL = "http://localhost:3000/books"
FILE_PATH = os.path.join(os.path.dirname(__file__), "books.json")

def send_books():
    # Lê o arquivo JSON
    with open(FILE_PATH, "r", encoding="utf-8") as f:
        books = json.load(f)

    for idx, book in enumerate(books, start=1):
        try:
            response = requests.post(API_URL, json=book)
            response.raise_for_status()
            print(f"[{idx}/{len(books)}] Livro enviado: {book['title']}")
        except requests.exceptions.RequestException as e:
            print(f"[ERRO] Falha ao enviar '{book['title']}': {e}")

if __name__ == "__main__":
    send_books()

