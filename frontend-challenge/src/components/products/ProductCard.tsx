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
import Image from 'next/image';
import { Minus, Plus, ShoppingCart } from 'lucide-react';
import { Product } from '@/types';
import { useCartStore, selectCartItem, MAX_QUANTITY } from '@/store/useCartStore';

interface ProductCardProps {
    product: Product;
}

export default function ProductCard({ product }: ProductCardProps) {
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
    const isTablet = useMediaQuery(theme.breakpoints.between('sm', 'md'));

    const cartItem = useCartStore(selectCartItem(product.id));
    const addItem = useCartStore((state) => state.addItem);
    const updateQuantity = useCartStore((state) => state.updateQuantity);
    const quantity = cartItem?.quantity || 0;

    const imageUrl = React.useMemo(() => {
        if (!product.image) return '';
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
                    mb: 3,
                    border: quantity > 0 ? `2px solid ${theme.palette.primary.main}` : 'none',
                    boxShadow: '0px 4px 20px rgba(0, 0, 0, 0.05)',
                    transition: 'transform 0.2s',
                    '&:hover': {
                        transform: 'translateY(-4px)',
                    },
                }}
            >
                {imageUrl ? (
                    <Image
                        src={imageUrl}
                        alt={product.name}
                        fill
                        sizes="(max-width: 600px) 100vw, (max-width: 900px) 50vw, 33vw"
                        style={{ objectFit: 'cover', borderRadius: 'inherit' }}
                    />
                ) : (
                    <Box sx={{ width: '100%', height: '100%', bgcolor: 'surface.main', borderRadius: 'inherit' }} />
                )}

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
                            aria-label={`Add ${product.name} to cart`}
                            sx={{
                                bgcolor: 'white',
                                color: theme.palette.text.primary,
                                border: `1px solid ${theme.palette.muted.main}`,
                                borderRadius: '999px',
                                width: '100%',
                                '&:hover': {
                                    bgcolor: 'surface.main',
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
                                aria-label={`Decrease quantity of ${product.name}`}
                                sx={{ color: 'white', border: '1px solid white', p: 0.5 }}
                            >
                                <Minus size={14} strokeWidth={3} />
                            </IconButton>
                            <Typography fontWeight={600}>{quantity}</Typography>
                            <IconButton
                                size="small"
                                onClick={() => updateQuantity(product.id, Math.min(quantity + 1, MAX_QUANTITY))}
                                aria-label={`Increase quantity of ${product.name}`}
                                disabled={quantity >= MAX_QUANTITY}
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
