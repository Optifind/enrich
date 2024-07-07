import { render, screen } from '@testing-library/react';
import App from './App';

test('renders header', () => {
    render(<App />);
    const header = screen.getByText(/Product Catalog/i);
    expect(header).toBeInTheDocument();
});

test('renders product display', () => {
    render(<App />);
    const display = screen.getByTestId("Product-display");
    expect(display).toBeInTheDocument();
});