<script lang="ts">
import { defineComponent, type PropType } from 'vue';

type Variant = 'h1' |
   'h2' |
   'h3' |
   'h4' |
   'h5' |
   'h6' |
   'h7' |
   'h8' |
   'span' |
   'buttonText' |
   'tiny' |
   'small' |
   'bodyText' |
   'subheading';


export default defineComponent({
  props: {
    variant: { type: Object as PropType<Variant> },
    upperCase: { type: Boolean },
    class: { type: String },
  },
  computed: {
    Component() {
      switch (this.variant) {
        case 'h1': return 'h1';
        case 'h2': return 'h2';
        case 'h3': return 'h3';
        case 'h4': return 'h4';
        case 'h5': return 'h5';
        case 'h6': return 'h6';
        case 'h7': return 'p';
        case 'h8': return 'p';
        case 'span': return 'span';
        case 'buttonText': return 'h5';

        default: return 'p';
      }
    },
    classes() {
      const classes = ['typography', this.variant, this.class];

      if (this.upperCase) {
        classes.push('upperCase');
      }

      return classes.join(' ');
    }

  }
});

</script>

<template>
  <component :is="Component" :class="classes">
    <slot></slot>
  </component>
</template>

<style lang="scss" scoped>

.typography {
  font-family: 'Source Sans Pro', sans-serif;
  margin: 0;
  color: $black-color;
}
.tiny {
  font-style: normal;
  font-weight: 300;
  font-size: 12px;
  line-height: 16px;
}
.small {
  font-style: normal;
  font-weight: 400;
  font-size: 12px;
  line-height: 16px;
}
.bodyText {
  font-style: normal;
  font-weight: 300;
  font-size: 16px;
  line-height: 24px;
}
.buttonText {
  font-family: 'Oswald';
  font-style: normal;
  font-weight: 400;
  font-size: 18px;
  line-height: 24px;
  color: $black-color;
  opacity: 0.5;
  display: inline;
}

.subheading {
  font-style: normal;
  font-weight: 700;
  font-size: 16px;
  line-height: 24px;
}
.h3 {
  font-family: 'Oswald';
  font-style: normal;
  font-weight: 400;
  font-size: 24px;
  line-height: 32px;
}
.h5 {
  font-family: 'Oswald';
  font-style: normal;
  font-weight: 400;
  font-size: 18px;
  line-height: 24px;
}
.upperCase {
  text-transform: uppercase;
}

</style>
