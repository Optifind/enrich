import React from 'react';
import { Product } from './../types';
import './componentsCSS/ProductBox1.css';


interface ProductBox1Props {
    product: Product;
    clickAction: (product: Product) => void;
}

const ProductBox1: React.FC<ProductBox1Props> = ({ product, clickAction }) => {
    const handleClick = (product: Product) => {
        clickAction(product);
    };

    return (
        <div className="product-box" onClick={() => handleClick(product)}>
            <h4>{product.name}</h4>
            <img src={product.image_url} alt={product.name} />
            <p>Price: ${product.price}</p>
        </div>
    );
};

export default ProductBox1;