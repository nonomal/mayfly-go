// 定义语言国际化内容
/**
 * 说明：
 * 注意国际化定义的字段，不要与原有的定义字段相同。
 * /src/i18n/(zh-cn、en...)/module.ts 下的 ts 为各模块国际化内容。
 */
import { getThemeConfig } from '@/common/utils/storage';
import { createI18n } from 'vue-i18n';

const modules: Record<string, any> = import.meta.glob('./**/*.ts', { eager: true });

function initI18n() {
    // 定义变量内容
    const messages: any = {};
    const itemizeMap = new Map<string, any[]>();

    // 对自动引入的 modules 进行分类 en、zh-cn
    // https://vitejs.cn/vite3-cn/guide/features.html#glob-import
    for (const path in modules) {
        const parts = path.split('/');
        const i18n = parts[1];

        const msgs = modules[path].default;
        if (itemizeMap.get(i18n)) {
            itemizeMap.get(i18n)?.push(modules[path].default);
        } else {
            itemizeMap.set(i18n, [msgs]);
        }
    }

    // 处理最终格式
    itemizeMap.forEach((value, key) => {
        messages[key] = Object.assign({}, ...value);
    });

    const themeConfig = getThemeConfig();
    const globalI18n = themeConfig.globalI18n || "zh-cn";

    // https://vue-i18n.intlify.dev/guide/essentials/fallback.html#explicit-fallback-with-one-locale
    return createI18n({
        legacy: false,
        globalInjection: true, // 在所有组件中都可以使用 $i18n $t $rt $d $n $tm
        silentTranslationWarn: true,
        missingWarn: false,
        silentFallbackWarn: true,
        fallbackWarn: false,
        locale: globalI18n,
        fallbackLocale: "zh-cn",
        messages,
    });
}

// 导出语言国际化
export const i18n: any = initI18n();
