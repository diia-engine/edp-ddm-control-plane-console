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
    case 'registryNameAlreadyExists':
      return 'Реєстр з такою назвою вже існує';
    case 'invalidFileType':
      return 'Невірний тип файлу';
    default:
      return 'Перевірте формат поля';
  }
};
