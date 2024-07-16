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
GET https://foo.bar/products?count={count}&similar_to={ids}&searchtype={searchtype}
```

`count` (optional) Number of products to retrieve. Defaults to 50 products if not specified.

`similar_to` (optional) Product IDs to query for similar products. Multiple IDs should be separated by commas (e.g., 123,124).

`searchtype` (optional) Criteria for similarity search, either tags or vectors. Defaults to backend-defined criteria if not specified.

### Response

#### Success

- **Code**: 200 OK
- **Content**: JSON array of products, sorted by **descending similarity** or **descending ID** if request contains no `similar_to` targets

```json
[
  {
    "id": "ID-420",
    "product-url": "https://store.com/product...",
    "product-image": "https://stor.com/image...",
    "title": "Dummy Product",
    "description": "This is a dummy product.",
    "price": "99,99 EUR"
  },
  {
    "id": "ID-69",
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
  - Description: No `ids` were provided even though `searchtype` was provided.
  - Description: `searchtype` was other than `vectors` or `tags`.
  - Description: `count` is not positive integer or null.
- **Code**: 404 Not found
  - Description: No products found for the given `ids`.
- **Code**: 204 No Content
  - Description: No products related close enough to be returned.
