'use client';

import React from 'react';
import {
    Dialog,
    DialogContent,
    Typography,
    Button,
    Stack,
    Box,
    Divider,
} from '@mui/material';
import Image from 'next/image';
import { CircleCheck } from 'lucide-react';
import { OrderResponse } from '@/types';

interface OrderModalProps {
    open: boolean;
    order: OrderResponse | null;
    onReset: () => void;
}

export default function OrderModal({ open, order, onReset }: OrderModalProps) {
    return (
        <Dialog
            open={open && !!order}
            maxWidth="xs"
            fullWidth
            aria-labelledby="order-confirm-title"
            PaperProps={{
                sx: { borderRadius: 4, p: 2 }
            }}
        >
            {order && (
                <DialogContent>
                    <Stack spacing={2} sx={{ mb: 3 }}>
                        <CircleCheck size={48} color="#1EA94C" />
                        <Box>
                            <Typography id="order-confirm-title" variant="h2" sx={{ color: 'text.primary', mb: 1 }}>
                                Order Confirmed
                            </Typography>
                            <Typography variant="body1" color="text.secondary">
                                We hope you enjoy your food!
                            </Typography>
                        </Box>
                    </Stack>

                    <Box sx={{ bgcolor: 'surface.main', px: 3, pt: 3, pb: 1, borderRadius: 2, mb: 4 }}>
                        <Stack spacing={2} sx={{ mb: 2 }}>
                            {order.items.map((item) => {
                                const p = order.products.find(prod => prod.id === item.productId);
                                if (!p) return null;
                                return (
                                    <Box key={item.productId}>
                                        <Stack direction="row" justifyContent="space-between" alignItems="center">
                                            <Stack direction="row" spacing={2} alignItems="center">
                                                <Box
                                                    sx={{
                                                        position: 'relative',
                                                        width: 48,
                                                        height: 48,
                                                        borderRadius: 1,
                                                        overflow: 'hidden',
                                                        flexShrink: 0,
                                                    }}
                                                >
                                                    {p.image?.thumbnail ? (
                                                        <Image
                                                            src={p.image.thumbnail}
                                                            alt={p.name}
                                                            fill
                                                            sizes="48px"
                                                            style={{ objectFit: 'cover' }}
                                                        />
                                                    ) : (
                                                        <Box sx={{ width: '100%', height: '100%', bgcolor: 'surface.main' }} />
                                                    )}
                                                </Box>
                                                <Box>
                                                    <Typography variant="body2" fontWeight={700} noWrap sx={{ maxWidth: '140px' }}>
                                                        {p.name}
                                                    </Typography>
                                                    <Stack direction="row" spacing={1}>
                                                        <Typography variant="body2" color="primary" fontWeight={700}>
                                                            {item.quantity}x
                                                        </Typography>
                                                        <Typography variant="body2" color="text.secondary">
                                                            @ ${p.price.toFixed(2)}
                                                        </Typography>
                                                    </Stack>
                                                </Box>
                                            </Stack>
                                            <Typography variant="body1" fontWeight={600}>
                                                ${(p.price * item.quantity).toFixed(2)}
                                            </Typography>
                                        </Stack>
                                        <Divider sx={{ mt: 2, borderColor: 'muted.main', opacity: 0.1 }} />
                                    </Box>
                                );
                            })}
                        </Stack>

                        <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ py: 2 }}>
                            <Typography variant="body1">Order Total</Typography>
                            <Typography variant="h2" color="text.primary">
                                ${order.total.toFixed(2)}
                            </Typography>
                        </Stack>

                        {order.discounts > 0 && (
                            <Typography variant="body2" color="primary.main" align="right" sx={{ mt: -1, mb: 1 }}>
                                Savings: -${order.discounts.toFixed(2)}
                            </Typography>
                        )}
                    </Box>

                    <Button
                        fullWidth
                        variant="contained"
                        size="large"
                        onClick={onReset}
                        sx={{ py: 2, fontSize: '18px' }}
                    >
                        Start New Order
                    </Button>
                </DialogContent>
            )}
        </Dialog>
    );
}
