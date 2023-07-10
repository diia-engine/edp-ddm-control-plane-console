export const getErrorMessage = (key: string, name?: string): string => {
  switch (key) {
    case 'required':
      return 'Поле обов’язкове для заповнення.';
    case 'moreThanMaxValue':
      return 'Перевищено максимально допустиме значення';
    case 'isUnique':
      return 'Неунікальне значення';
    default:
      if (name && key !== `${name} is invalid`) {
        return key;
      }
      return 'Перевірте формат поля';
  }
};
