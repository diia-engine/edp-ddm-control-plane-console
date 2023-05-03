export const getErrorMessage = (key: string): string => {
    switch (key) {
        case 'required':
            return 'Не може бути порожнім';
        case 'only-integer':
            return 'Тільки цілі числа';
        case 'cron-expression':
            return 'Невірний вираз';
        case 'checkFormat':
          return 'Перевірте формат поля';
        case 'invalidFormat':
          return 'Невірний формат';
        default:
            return '';
    }
};
