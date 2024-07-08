import React from 'react';
import { Product } from './../types';
import './componentsCSS/Recommendations.css';

type RecommendationsProps = {
    productIds: number[];
    products: Product[];
};

const Recommendations: React.FC<RecommendationsProps> = ({ products }) => {
    return (
        <div className='recommended-products'>
            <h2>Recommended Products</h2>
        </div>
    );
};

export default Recommendations;