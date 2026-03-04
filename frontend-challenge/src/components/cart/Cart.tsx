'use client';

import React from 'react';
import {
    Box,
    Typography,
    Button,
    IconButton,
    Stack,
    Divider,
    Paper,
    TextField,
    Dialog,
    DialogTitle,
    DialogContent,
    DialogContentText,
    DialogActions,
    CircularProgress
} from '@mui/material';
import { XCircle, BadgeCheck, UtensilsCrossed } from 'lucide-react';
import { useCartStore, selectTotalItems, selectSubtotal } from '@/store/useCartStore';

interface CartProps {
    onConfirm: (couponCode: string) => void;
    promoError?: string | null;
    onPromoErrorClear?: () => void;
    isPending?: boolean;
}

export default function Cart({ onConfirm, promoError, onPromoErrorClear, isPending }: CartProps) {
    const items = useCartStore((state) => state.items);
    const removeItem = useCartStore((state) => state.removeItem);
    const totalItems = useCartStore(selectTotalItems);
    const subtotal = useCartStore(selectSubtotal);

    const [couponCode, setCouponCode] = React.useState('');
    const [deleteItemId, setDeleteItemId] = React.useState<string | null>(null);

    React.useEffect(() => {
        if (totalItems === 0) {
            setCouponCode('');
        }
    }, [totalItems]);

    const handleCouponChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setCouponCode(e.target.value);
        if (promoError && onPromoErrorClear) {
            onPromoErrorClear();
        }
    };

    const handleDeleteClick = (productId: string) => {
        setDeleteItemId(productId);
    };

    const handleConfirmDelete = () => {
        if (deleteItemId) {
            removeItem(deleteItemId);
            setDeleteItemId(null);
        }
    };

    const handleCancelDelete = () => {
        setDeleteItemId(null);
    };

    if (totalItems === 0) {
        return (
            <Paper elevation={0} sx={{ p: 4, borderRadius: 3, bgcolor: 'white' }}>
                <Typography variant="h2" color="primary" sx={{ mb: 3 }}>
                    Your Cart ({totalItems})
                </Typography>
                <Stack alignItems="center" spacing={2} sx={{ py: 4 }}>
                    <UtensilsCrossed size={120} color="currentColor" strokeWidth={1} />
                    <Typography color="text.secondary" fontWeight={600}>
                        Your added items will appear here
                    </Typography>
                </Stack>
            </Paper>
        );
    }

    return (
        <Paper elevation={0} sx={{ p: 4, borderRadius: 3, bgcolor: 'white' }}>
            <Typography variant="h2" color="primary" sx={{ mb: 3 }}>
                Your Cart ({totalItems})
            </Typography>

            <Stack spacing={2} sx={{ mb: 3 }}>
                {items.map((item) => (
                    <Box key={item.productId}>
                        <Stack direction="row" justifyContent="space-between" alignItems="center">
                            <Box>
                                <Typography variant="body1" fontWeight={600} sx={{ mb: 0.5 }}>
                                    {item.product.name}
                                </Typography>
                                <Stack direction="row" spacing={2} alignItems="center">
                                    <Typography variant="body2" color="primary" fontWeight={700}>
                                        {item.quantity}x
                                    </Typography>
                                    <Typography variant="body2" color="text.secondary">
                                        @ ${item.product.price.toFixed(2)}
                                    </Typography>
                                    <Typography variant="body2" color="text.secondary" fontWeight={600}>
                                        ${(item.product.price * item.quantity).toFixed(2)}
                                    </Typography>
                                </Stack>
                            </Box>
                            <IconButton
                                size="small"
                                onClick={() => handleDeleteClick(item.productId)}
                                aria-label={`Remove ${item.product.name} from cart`}
                                sx={{ color: 'muted.main' }}
                            >
                                <XCircle size={20} />
                            </IconButton>
                        </Stack>
                        <Divider sx={{ mt: 2, borderColor: 'surface.main' }} />
                    </Box>
                ))}
            </Stack>

            <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 3 }}>
                <Typography variant="body1">Order Total</Typography>
                <Typography variant="h2" color="text.primary">
                    ${subtotal.toFixed(2)}
                </Typography>
            </Stack>

            <Box sx={{ mb: 3 }}>
                <TextField
                    fullWidth
                    size="small"
                    placeholder="Promo Code"
                    value={couponCode}
                    error={!!promoError}
                    helperText={promoError}
                    onChange={handleCouponChange}
                    sx={{
                        '& .MuiOutlinedInput-root': {
                            borderRadius: 2,
                            bgcolor: promoError ? 'surface.light' : 'surface.main',
                            '& fieldset': { borderColor: promoError ? 'error.main' : 'transparent' },
                            '&:hover fieldset': { borderColor: promoError ? 'error.main' : 'primary.main' },
                        },
                        '& .MuiFormHelperText-root': {
                            mx: 0,
                            fontWeight: 600,
                        }
                    }}
                />
            </Box>

            <Box sx={{ bgcolor: 'surface.main', p: 2, borderRadius: 2, mb: 3 }}>
                <Stack direction="row" spacing={1} alignItems="center">
                    <BadgeCheck size={20} color="currentColor" />
                    <Typography variant="body2">
                        This is a <strong>carbon-neutral</strong> delivery
                    </Typography>
                </Stack>
            </Box>

            <Button
                fullWidth
                variant="contained"
                size="large"
                onClick={() => onConfirm(couponCode)}
                disabled={isPending}
                sx={{ py: 2, fontSize: '18px' }}
            >
                {isPending ? <CircularProgress size={24} color="inherit" /> : 'Confirm Order'}
            </Button>

            {/* Deletion Confirmation Dialog */}
            <Dialog
                open={!!deleteItemId}
                onClose={handleCancelDelete}
                PaperProps={{
                    sx: { borderRadius: 3, p: 1 }
                }}
            >
                <DialogTitle sx={{ fontWeight: 700 }}>Remove Item?</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        Are you sure you want to remove this item from your cart?
                    </DialogContentText>
                </DialogContent>
                <DialogActions sx={{ px: 3, pb: 2 }}>
                    <Button onClick={handleCancelDelete} color="inherit" sx={{ fontWeight: 600 }}>
                        Cancel
                    </Button>
                    <Button onClick={handleConfirmDelete} variant="contained" color="primary" autoFocus sx={{ borderRadius: '999px', px: 4 }}>
                        Remove
                    </Button>
                </DialogActions>
            </Dialog>
        </Paper>
    );
}
