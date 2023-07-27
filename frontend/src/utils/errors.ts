export const getErrorMessage = (key: string): string => {
  switch (key) {
    case 'required':
      return 'Поле обов’язкове для заповнення.';
    case 'moreThanMaxValue':
      return 'Перевищено максимально допустиме значення';
    case 'isUnique':
      return 'Неунікальне значення';
    case 'rateLimitError':
      return 'Вкажіть ліміт мінімум в одному полі';
    default:
      return 'Перевірте формат поля';
  }
};
