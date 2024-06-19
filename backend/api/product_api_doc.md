# Routes

---

## Get single product

### Request

```http request
GET https://foo.bar/product/{id}
```

`id` (required) The product ID to query.

### Response

#### Success

- **Code**: 200 OK
- **Content**: JSON with product information

```json
{
  "id": 123,
  "product-url": "https://store.com/product...",
  "product-image": "https://store.com/image...",
  "title": "Dummy Product",
  "description": "This is a dummy product.",
  "price": "99,99 EUR"
}
```

#### Errors

- **Code**: 404 Not found
    - Description: No product found for the given `id`.
- **Code**: 400 Bad request
    - Description: No `id` was provided.

---

## Get multiple products

### Request

```http request
GET https://foo.bar/products?count={count}
```

`count` (optional) Amount of products to retrieve. If count is left empty the API will return all products in database. If count exceeds the number of products in database, all products are returned.

### Response

#### Success

- **Code**: 200 OK
- **Content**: JSON array of products, sorted by **descending ID**

```json
[
  {
    "id": 123,
    "product-url": "https://store.com/product...",
    "product-image": "https://stor.com/image...",
    "title": "Dummy Product",
    "description": "This is a dummy product.",
    "price": "99,99 EUR"
  },
  {
    "id": 124,
    "product-url": "https://store.com/product...",
    "product-image": "https://store.com/image...",
    "title": "Another Product",
    "description": "This is some other dummy product.",
    "price": "11,11 EUR"
  }
]
```

#### Errors

- **Code**: 400 Bad request
  - Description: `count` is not positive integer or null.

---

## Get similar products

### Request

```http request
GET https://foo.bar/products/{id}/similar?count={count}
```

`id` (required) The product ID to query.

`count` (optional) Amount of products to retrieve. If count is left empty the API will return all products in database. If count exceeds the number of products in database, all products are returned.

### Response

#### Success

- **Code**: 200 OK
- **Content**: JSON array of products, sorted by **descending similarity**

```json
[
  {
    "id": 123,
    "product-url": "https://store.com/product...",
    "product-image": "https://store.com/image...",
    "title": "Queried Product",
    "description": "This is the product that was queried.",
    "price": "99,99 EUR"
  },
  {
    "id": 124,
    "product-url": "https://store.com/product...",
    "product-image": "https://store.com/image...",
    "title": "Similar Product",
    "description": "This is a product that is the most similar to the one queried.",
    "price": "11,11 EUR"
  }
]
```

#### Errors

- **Code**: 400 Bad request
    - Description: `count` is not positive integer or null.
- **Code**: 400 Bad request
    - Description: No `id` was provided.
- **Code**: 404 Not found
    - Description: No product found for the given `id`.
