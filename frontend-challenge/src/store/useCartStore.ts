import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { Product, CartItem } from '../types';

interface CartState {
    items: CartItem[];
    addItem: (product: Product) => void;
    removeItem: (productId: string) => void;
    updateQuantity: (productId: string, quantity: number) => void;
    clearCart: () => void;
    getTotalItems: () => number;
    getSubtotal: () => number;
    syncPrices: (currentProducts: Product[]) => string[];
}

export const useCartStore = create<CartState>()(
    persist(
        (set, get) => ({
            items: [],

            addItem: (product: Product) => {
                set((state) => {
                    const existingItem = state.items.find((item) => item.productId === product.id);
                    if (existingItem) {
                        return {
                            items: state.items.map((item) =>
                                item.productId === product.id
                                    ? { ...item, quantity: item.quantity + 1 }
                                    : item
                            ),
                        };
                    }
                    return {
                        items: [...state.items, { productId: product.id, quantity: 1, product }],
                    };
                });
            },

            removeItem: (productId: string) => {
                set((state) => ({
                    items: state.items.filter((item) => item.productId !== productId),
                }));
            },

            updateQuantity: (productId: string, quantity: number) => {
                if (quantity <= 0) {
                    get().removeItem(productId);
                    return;
                }
                set((state) => ({
                    items: state.items.map((item) =>
                        item.productId === productId ? { ...item, quantity } : item
                    ),
                }));
            },

            clearCart: () => set({ items: [] }),

            getTotalItems: () => {
                return get().items.reduce((total, item) => total + item.quantity, 0);
            },

            getSubtotal: () => {
                return get().items.reduce(
                    (total, item) => total + item.product.price * item.quantity,
                    0
                );
            },

            syncPrices: (currentProducts: Product[]) => {
                let warnings: string[] = [];
                set((state) => {
                    let hasChanges = false;
                    const newItems = state.items.map((item) => {
                        const currentProduct = currentProducts.find(p => p.id === item.productId);

                        // Item was deleted or went out of stock
                        if (!currentProduct) {
                            hasChanges = true;
                            warnings.push(`${item.product.name} is no longer available and was removed from your cart.`);
                            return null;
                        }

                        // Price increased
                        if (currentProduct.price > item.product.price) {
                            hasChanges = true;
                            warnings.push(`The price of ${currentProduct.name} has increased to $${currentProduct.price.toFixed(2)}.`);
                            return { ...item, product: currentProduct };
                        }

                        // Price decreased or other silent product update
                        if (currentProduct.price !== item.product.price || currentProduct.name !== item.product.name) {
                            hasChanges = true;
                            return { ...item, product: currentProduct };
                        }

                        return item;
                    }).filter(Boolean) as CartItem[]; // Remove nulls (deleted items)

                    return hasChanges ? { items: newItems } : state;
                });
                return warnings;
            },
        }),
        {
            name: 'shopping-cart-storage',
        }
    )
);
