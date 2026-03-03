'use client';

import React from 'react';
import {
    Box,
    Typography,
    Button,
    IconButton,
    Stack,
    useTheme,
    useMediaQuery
} from '@mui/material';
import { Minus, Plus, ShoppingCart } from 'lucide-react';
import { Product } from '@/types';
import { useCartStore } from '@/store/useCartStore';

interface ProductCardProps {
    product: Product;
}

export default function ProductCard({ product }: ProductCardProps) {
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
    const isTablet = useMediaQuery(theme.breakpoints.between('sm', 'md'));

    const { items, addItem, updateQuantity } = useCartStore();
    const cartItem = items.find((item) => item.productId === product.id);
    const quantity = cartItem?.quantity || 0;

    const imageUrl = React.useMemo(() => {
        if (!product.image) return '/placeholder.jpg';
        if (isMobile) return product.image.mobile;
        if (isTablet) return product.image.tablet;
        return product.image.desktop;
    }, [product.image, isMobile, isTablet]);

    return (
        <Box sx={{ width: '100%', mb: 4 }}>
            <Box
                sx={{
                    position: 'relative',
                    width: '100%',
                    aspectRatio: '1/1',
                    borderRadius: 2,
                    // overflow: 'hidden', // Removed to prevent clipping the 'Add to Cart' button
                    mb: 3,
                    border: quantity > 0 ? `2px solid ${theme.palette.primary.main}` : 'none',
                    boxShadow: '0px 4px 20px rgba(0, 0, 0, 0.05)',
                    transition: 'transform 0.2s',
                    '&:hover': {
                        transform: 'translateY(-4px)',
                    },
                }}
            >
                <img
                    src={imageUrl}
                    alt={product.name}
                    style={{ width: '100%', height: '100%', objectFit: 'cover' }}
                />

                {/* Add to Cart Overlay */}
                <Box
                    sx={{
                        position: 'absolute',
                        bottom: -24,
                        left: '50%',
                        transform: 'translateX(-50%)',
                        width: '160px',
                        zIndex: 10,
                    }}
                >
                    {quantity === 0 ? (
                        <Button
                            variant="contained"
                            color="inherit"
                            startIcon={<ShoppingCart size={20} color={theme.palette.primary.main} />}
                            onClick={() => addItem(product)}
                            sx={{
                                bgcolor: 'white',
                                color: theme.palette.text.primary,
                                border: '1px solid #AD8982',
                                borderRadius: '999px',
                                width: '100%',
                                '&:hover': {
                                    bgcolor: '#F5F5F5',
                                    borderColor: theme.palette.primary.main,
                                },
                            }}
                        >
                            Add to Cart
                        </Button>
                    ) : (
                        <Stack
                            direction="row"
                            alignItems="center"
                            justifyContent="space-between"
                            sx={{
                                bgcolor: theme.palette.primary.main,
                                color: 'white',
                                borderRadius: '999px',
                                height: '44px',
                                px: 1,
                            }}
                        >
                            <IconButton
                                size="small"
                                onClick={() => updateQuantity(product.id, quantity - 1)}
                                sx={{ color: 'white', border: '1px solid white', p: 0.5 }}
                            >
                                <Minus size={14} strokeWidth={3} />
                            </IconButton>
                            <Typography fontWeight={600}>{quantity}</Typography>
                            <IconButton
                                size="small"
                                onClick={() => updateQuantity(product.id, quantity + 1)}
                                sx={{ color: 'white', border: '1px solid white', p: 0.5 }}
                            >
                                <Plus size={14} strokeWidth={3} />
                            </IconButton>
                        </Stack>
                    )}
                </Box>
            </Box>

            <Typography variant="body2" color="text.secondary" sx={{ mb: 0.5 }}>
                {product.category}
            </Typography>
            <Typography variant="body1" fontWeight={600} sx={{ mb: 0.5 }}>
                {product.name}
            </Typography>
            <Typography variant="body1" fontWeight={700} color="primary">
                ${product.price.toFixed(2)}
            </Typography>
        </Box>
    );
}
