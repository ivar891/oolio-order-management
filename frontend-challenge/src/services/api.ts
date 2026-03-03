import axios from 'axios';
import { OrderRequest, OrderResponse, Product } from '../types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const API_KEY = process.env.NEXT_PUBLIC_API_KEY || 'apitest';

const api = axios.create({
    baseURL: API_BASE_URL,
    headers: {
        'Content-Type': 'application/json',
        'api_key': API_KEY,
    },
});

export const getProducts = async (): Promise<Product[]> => {
    const response = await api.get<Product[]>('/api/product');
    return response.data;
};

export const placeOrder = async (order: OrderRequest): Promise<OrderResponse> => {
    const response = await api.post<OrderResponse>('/api/order', order);
    return response.data;
};

export default api;
