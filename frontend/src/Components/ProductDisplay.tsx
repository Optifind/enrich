import React from 'react';
import { Product } from './../types';

//placeholder data
const products: Product[] = [
    { id: 1, name: 'Product 1', price: 10, description: 'This is a description of Product 1'},
    { id: 2, name: 'Product 2', price: 20, description: 'This is a description of Product 2'},
    { id: 3, name: 'Product 3', price: 30, description: 'This is a description of Product 3'},
    { id: 4, name: 'Product 4', price: 40, description: 'This is a description of Product 4'},
    { id: 5, name: 'Product 5', price: 50, description: 'This is a description of Product 5'},
    { id: 6, name: 'Product 6', price: 60, description: 'This is a description of Product 6'},
];

const ProductDisplay: React.FC = () => {
    return (
        <div>
            <h2>Product Display</h2>
            {products.map((product) => (
                <div key={product.id}>
                    <h3>{product.name}</h3>
                    <p>Price: ${product.price}</p>
                </div>
            ))}
        </div>
    );
};

export default ProductDisplay;