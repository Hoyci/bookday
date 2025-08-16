import requests
import random
import json
import os
from stdnum import isbn as isbnlib  # pip install python-stdnum

def is_valid_isbn(value):
    """Valida se o ISBN é ISBN-10 ou ISBN-13."""
    try:
        isbnlib.validate(value)
        return True
    except Exception:
        return False

def generate_random_isbn13():
    """Gera um ISBN-13 válido."""
    prefix = "978"  # Prefixo comum
    base = prefix + "".join(str(random.randint(0, 9)) for _ in range(9))
    # Calcula dígito verificador
    total = sum((int(d) if i % 2 == 0 else int(d) * 3) for i, d in enumerate(base))
    check_digit = (10 - (total % 10)) % 10
    return base + str(check_digit)

def fetch_books(query="book", total_needed=100):
    books = []
    page = 1

    while len(books) < total_needed:
        params = {
            "q": query,
            "page": page,
            "limit": 100
        }
        response = requests.get("https://openlibrary.org/search.json", params=params)
        response.raise_for_status()
        data = response.json()

        works = data.get("docs", [])
        random.shuffle(works)

        for work in works:
            title = work.get("title")
            authors = work.get("author_name", [])
            isbns = work.get("isbn", [])
            ia_list = work.get("ia", [])

            if not title or not authors:
                continue

            isbn = None
            # 1️⃣ ISBN direto
            if isbns:
                for i in isbns:
                    if is_valid_isbn(i):
                        isbn = i
                        break

            # 2️⃣ ISBN no campo "ia"
            if not isbn and ia_list:
                for ia in ia_list:
                    if ia.startswith("isbn_"):
                        candidate = ia.replace("isbn_", "")
                        if is_valid_isbn(candidate):
                            isbn = candidate
                            break

            # 3️⃣ Gerar ISBN-13 válido se não encontrar
            if not isbn:
                isbn = generate_random_isbn13()

            book = {
                "title": title,
                "author": authors[0],
                "isbn": isbn,
                "catalog_price": round(random.uniform(10.0, 100.0), 2),
                "initial_stock": random.randint(1, 20)
            }
            books.append(book)

            if len(books) >= total_needed:
                break

        page += 1

    return books

if __name__ == "__main__":
    random_books = fetch_books(total_needed=100)
    file_path = os.path.join(os.path.dirname(__file__), "books.json")

    with open(file_path, "w", encoding="utf-8") as f:
        json.dump(random_books, f, indent=4, ensure_ascii=False)

    print(f"Arquivo salvo em: {file_path} ({len(random_books)} livros)")

