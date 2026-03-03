export interface ProductImage {
    thumbnail: string;
    mobile: string;
    tablet: string;
    desktop: string;
}

export interface Product {
    id: string;
    name: string;
    price: number;
    currency: string;
    category: string;
    image?: ProductImage;
}

export interface OrderItem {
    productId: string;
    quantity: number;
}

export interface OrderRequest {
    couponCode?: string;
    items: OrderItem[];
}

export interface OrderResponse {
    id: string;
    items: OrderItem[];
    products: Product[];
    couponCode?: string;
    currency: string;
    total: number;
    discounts: number;
}

export interface CartItem extends OrderItem {
    product: Product;
}
