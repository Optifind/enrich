import React from 'react';
import './App.css';
import ProductDisplay from './Components/ProductDisplay.tsx';
import Recommendations from './Components/Recommendations.tsx'

function App() {
    return (
        <div className="App">
            <header className="App-header">
                <h1 className="Header">Product Catalog</h1>
            </header>
            <ProductDisplay test-id="Product-display" />
            <Recommendations test-id="Recommendations" />
        </div>
    );
}

export default App;
