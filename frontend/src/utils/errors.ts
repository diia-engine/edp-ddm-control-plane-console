export const getErrorMessage = (key: string): string => {
    switch (key) {
        case 'required':
            return 'Не може бути порожнім';
        case 'only-integer':
            return 'Тільки цілі числа';  
        case 'cron-expression':
            return 'Невірний вираз';      
        default:
            return '';
    }
};