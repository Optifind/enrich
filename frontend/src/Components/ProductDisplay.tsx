import React, { useState } from 'react';
import ProductBox1 from './ProductBox1.tsx';
import { Product } from './../types';
import './componentsCSS/ProductDisplay.css';

//placeholder data
import { products } from './../API/placeholder.ts';

const ProductDisplay = () => {
    const [selectedProducts, setSelectedProducts] = useState<Product[]>([]);

    function clickAction (product: Product) {
        if (!selectedProducts.some(p => p.id === product.id)) {
            const newSelectedProducts = [...selectedProducts];
            newSelectedProducts.push(product);
            setSelectedProducts(newSelectedProducts);
            console.log('Selected Products:', newSelectedProducts);
        }
    }

    return (
        <div className='display'>
            <h3>Product Display</h3>
            <div className="container">
                <ProductBox1 product={products[0]} clickAction={clickAction} />
                <ProductBox1 product={products[1]} clickAction={clickAction} />
                <ProductBox1 product={products[2]} clickAction={clickAction} />
            </div>
        </div>
    );
};

export default ProductDisplay;