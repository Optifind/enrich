import React, { useState } from 'react';
import './App.css';
import ProductDisplay from './Components/ProductDisplay.tsx';
import Recommendations from './Components/Recommendations.tsx'
import { recommendedProducts } from './API/placeholder.ts';

function App() {
    const [showRecommendations, setShowRecommendations] = useState(false);

    return (
        <div className="App">
            <header className="App-header">
                <h1 className="Header">Optifind Enrich Demo</h1>
            </header>
            <ProductDisplay test-id="Product-display" />
            <button onClick={() => setShowRecommendations(!showRecommendations)}>
                {showRecommendations ? 'Hide Recommendations' : 'Show Recommendations'}
            </button>
            {showRecommendations && <Recommendations recommendedProducts={recommendedProducts} test-id="Recommendations" />}
        </div>
    );
}

export default App;
