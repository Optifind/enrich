import React from 'react';
import { Product } from './../types';

type RecommendationsProps = {
    productIds: number[];
    products: Product[];
};

const Recommendations: React.FC<RecommendationsProps> = ({ productIds, products }) => {
    const recommendedProducts = products.filter((product) => productIds.includes(product.id)).slice(0, 3);

    return (
        <div>
            <h2>Recommended Products</h2>
            <ul>
                {recommendedProducts.map((product) => (
                    <li key={product.id}>
                        <h3>{product.name}</h3>
                        <p>Price: ${product.price}</p>
                    </li>
                ))}
            </ul>
        </div>
    );
};

export default Recommendations;