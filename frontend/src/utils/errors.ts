export const getErrorMessage = (key: string): string => {
    switch (key) {
        case 'required':
            return 'Поле обов’язкове для заповнення';
        case 'only-integer':
            return 'Тільки цілі числа';
        case 'cron-expression':
            return 'Невірний вираз';
        case 'checkFormat':
          return 'Перевірте формат поля';
        case 'invalidFormat':
          return 'Невірний формат';
        case 'moreThanMaxValue': 
            return 'Перевищено максимально допустиме значення';
        default:
            return '';
    }
};
