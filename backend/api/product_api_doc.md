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

## Get similar products to one product

### Request

```http request
GET https://foo.bar/products/similar_to={id}?count={count}
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
    - Description: No `id` was provided.
- **Code**: 400 Bad request
    - Description: `count` is not positive integer or null.
- **Code**: 404 Not found
    - Description: No product found for the given `id`.
- **Code**: 204 No Content
    - Description: No product related close enought to be returned.

## Get similar products to multiple products

We need to be able to combine information from multiple items so that our results truly find a style.

### Request

```http request
GET https://foo.bar/products/similar_to={ids_separated_by_underscores}?searchtype={searchtype}?count={count}
```

`ids_separated_by_underscores` (required) The product IDs to query. We can get these from products the user has clicked on the moodboard or products they have previously bought. Backend can distinguish between one/multiple products being searched.

We need to process the products according to the id's. We create a median vector for the general style, and find vectors nearest to it.

`searchtype` (optional) Either `tags` or `vectors` to signal what are we searching with. Alternatively (or when left empty), we can decide this at the backend level.

Note: Tags that the products have in common could also be used to e.g. recommend pants that go well with a jacket that was searched, not just products of similar style.

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
    - Description: No `ids_separated_by_underscore` was provided.
- **Code**: 400 Bad request
    - Description: No `searchtype` was other than vectors or tags.
- **Code**: 400 Bad request
    - Description: `count` is not positive integer or null.
- **Code**: 404 Not found
    - Description: No product found for the given `id`.
- **Code**: 204 No Content
    - Description: No products related close enought to be returned.