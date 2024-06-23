import React from 'react';
import { Product } from '../types';

type ProductProps = {
    product: Product;
};

const Product: React.FC<ProductProps> = ({ product }) => {
    return (
        <div className="product">
            <img src={product.image_url} alt={product.name} />
            <h2>{product.name}</h2>
            <p>{product.description}</p>
            <p>Price: ${product.price.toFixed(2)}</p>
        </div>
    );
};

export default Product;
