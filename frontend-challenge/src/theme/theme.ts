import { createTheme } from '@mui/material/styles';

declare module '@mui/material/styles' {
    interface Palette {
        muted: Palette['primary'];
        surface: { main: string; light: string };
    }
    interface PaletteOptions {
        muted?: PaletteOptions['primary'];
        surface?: { main: string; light: string };
    }
}

export const getTheme = (fontFamily: string) =>
    createTheme({
        palette: {
            primary: {
                main: '#C73B0F',
                contrastText: '#FFFFFF',
            },
            secondary: {
                main: '#261108',
            },
            success: {
                main: '#1EA94C',
            },
            background: {
                default: '#F5F5F5',
                paper: '#FFFFFF',
            },
            text: {
                primary: '#261108',
                secondary: '#87635A',
            },
            muted: {
                main: '#AD8982',
                light: '#C9ADA7',
                dark: '#87635A',
                contrastText: '#FFFFFF',
            },
            surface: {
                main: '#F5F5F5',
                light: '#FFF5F5',
            },
        },
        typography: {
            fontFamily: fontFamily,
            h1: {
                fontSize: '40px',
                fontWeight: 700,
                lineHeight: '120%',
            },
            h2: {
                fontSize: '24px',
                fontWeight: 700,
            },
            body1: {
                fontSize: '16px',
                fontWeight: 400,
            },
            body2: {
                fontSize: '14px',
                fontWeight: 400,
            },
        },
        shape: {
            borderRadius: 8,
        },
        components: {
            MuiButton: {
                styleOverrides: {
                    root: {
                        textTransform: 'none',
                        borderRadius: '999px',
                        fontWeight: 600,
                        padding: '12px 24px',
                    },
                    containedPrimary: {
                        '&:hover': {
                            backgroundColor: '#A82F0A',
                        },
                    },
                },
            },
        },
    });
