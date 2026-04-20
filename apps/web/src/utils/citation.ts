import CSL from 'citeproc'

export type CSLStyleId = 
  | 'apa'
  | 'mla'
  | 'chicago'
  | 'ieee'
  | 'gb-t-7714-2015'
  | 'harvard-cite-them-right'
  | 'vancouver'
  | 'nature'
  | 'science'
  | 'acm-siggraph'

export interface CSLStyleInfo {
  id: CSLStyleId
  name: string
  nameEn: string
}

export const CSL_STYLES: CSLStyleInfo[] = [
  { id: 'apa', name: 'APA (美国心理学会)', nameEn: 'APA' },
  { id: 'mla', name: 'MLA (现代语言协会)', nameEn: 'MLA' },
  { id: 'chicago', name: '芝加哥格式', nameEn: 'Chicago' },
  { id: 'ieee', name: 'IEEE', nameEn: 'IEEE' },
  { id: 'gb-t-7714-2015', name: 'GB/T 7714-2015 (中国国家标准)', nameEn: 'GB/T 7714-2015' },
  { id: 'harvard-cite-them-right', name: '哈佛格式', nameEn: 'Harvard' },
  { id: 'vancouver', name: '温哥华格式', nameEn: 'Vancouver' },
  { id: 'nature', name: 'Nature', nameEn: 'Nature' },
  { id: 'science', name: 'Science', nameEn: 'Science' },
  { id: 'acm-siggraph', name: 'ACM SIGGRAPH', nameEn: 'ACM SIGGRAPH' },
]

const CSL_STYLE_BASE_URL = 'https://raw.githubusercontent.com/citation-style-language/styles/master/'

export class CitationFormatter {
  private styleCache: Map<string, string> = new Map()
  private sys: CSL.System

  constructor() {
    this.sys = {
      retrieveLocale: (lang: string) => {
        const locales: Record<string, string> = {
          'zh-CN': this.getZhLocale(),
          'en-US': this.getEnLocale(),
        }
        return locales[lang] || locales['en-US']
      },
      retrieveItem: () => {
        return {} as CSL.Item
      },
    }
  }

  async loadStyle(styleId: CSLStyleId): Promise<string> {
    if (this.styleCache.has(styleId)) {
      return this.styleCache.get(styleId)!
    }

    try {
      const response = await fetch(`${CSL_STYLE_BASE_URL}${styleId}.csl`)
      if (!response.ok) {
        throw new Error(`Failed to load style: ${styleId}`)
      }
      const styleXml = await response.text()
      this.styleCache.set(styleId, styleXml)
      return styleXml
    } catch (error) {
      console.error('Failed to load CSL style:', error)
      return this.getFallbackStyle(styleId)
    }
  }

  async formatCitation(
    items: CSL.Item[],
    styleId: CSLStyleId = 'apa',
    lang: string = 'en-US'
  ): Promise<string> {
    const styleXml = await this.loadStyle(styleId)
    
    const itemMap: Record<string, CSL.Item> = {}
    items.forEach((item, index) => {
      if (!item.id) {
        item.id = `item-${index}`
      }
      itemMap[item.id] = item
    })

    const sys: CSL.System = {
      retrieveLocale: (l: string) => {
        const locales: Record<string, string> = {
          'zh-CN': this.getZhLocale(),
          'en-US': this.getEnLocale(),
        }
        return locales[l] || locales['en-US']
      },
      retrieveItem: (id: string) => itemMap[id],
    }

    const citeproc = new CSL.Engine(sys, styleXml, lang, false)
    const ids = items.map(item => item.id)
    
    citeproc.updateItems(ids, false)
    const result = citeproc.makeBibliography()
    
    if (result && result[1].length > 0) {
      return result[1].join('\n')
    }
    
    return ''
  }

  async formatInline(
    item: CSL.Item,
    styleId: CSLStyleId = 'apa',
    lang: string = 'en-US'
  ): Promise<string> {
    const styleXml = await this.loadStyle(styleId)
    
    if (!item.id) {
      item.id = 'temp-item'
    }

    const sys: CSL.System = {
      retrieveLocale: (l: string) => {
        const locales: Record<string, string> = {
          'zh-CN': this.getZhLocale(),
          'en-US': this.getEnLocale(),
        }
        return locales[l] || locales['en-US']
      },
      retrieveItem: (id: string) => id === item.id ? item : ({} as CSL.Item),
    }

    const citeproc = new CSL.Engine(sys, styleXml, lang, false)
    const citation = {
      citationItems: [{ id: item.id }],
      properties: {},
    }
    
    const result = citeproc.appendCitationCluster(citation, 'cite')
    return result[0]?.[1] || ''
  }

  private getFallbackStyle(styleId: CSLStyleId): string {
    const styles: Record<string, string> = {
      'apa': this.getAPAStyle(),
      'mla': this.getMLAStyle(),
      'ieee': this.getIEEEStyle(),
      'gb-t-7714-2015': this.getGBT7714Style(),
    }
    return styles[styleId] || styles['apa']
  }

  private getAPAStyle(): string {
    return `<?xml version="1.0" encoding="utf-8"?>
<style xmlns="http://purl.org/net/xbiblio/csl" class="in-text" version="1.0" default-locale="en-US">
  <info><title>APA</title><id>apa</id></info>
  <citation et-al-min="3" et-al-use-first="1">
    <layout prefix="(" suffix=")" delimiter="; ">
      <text variable="author" suffix=", "/>
      <date variable="issued"><date-part name="year"/></date>
      <text variable="locator" prefix=": "/>
    </layout>
  </citation>
  <bibliography>
    <layout>
      <text variable="author" suffix=" "/>
      <date variable="issued" prefix="(" suffix="). "><date-part name="year"/></date>
      <text variable="title" suffix=". "/>
      <text variable="container-title" font-style="italic" suffix=", "/>
      <text variable="volume" suffix="("/>
      <text variable="issue" suffix="), "/>
      <text variable="page" suffix="."/>
    </layout>
  </bibliography>
</style>`
  }

  private getMLAStyle(): string {
    return `<?xml version="1.0" encoding="utf-8"?>
<style xmlns="http://purl.org/net/xbiblio/csl" class="in-text" version="1.0">
  <info><title>MLA</title><id>mla</id></info>
  <citation et-al-min="3" et-al-use-first="1">
    <layout prefix="(" suffix=")" delimiter="; ">
      <text variable="author"/>
      <text variable="locator" prefix=" "/>
    </layout>
  </citation>
  <bibliography>
    <layout>
      <text variable="author" suffix=". "/>
      <text variable="title" suffix=". " quotes="true"/>
      <text variable="container-title" suffix=", " font-style="italic"/>
      <text variable="volume" suffix=", "/>
      <text variable="issue" suffix=", "/>
      <date variable="issued"><date-part name="year"/></date>
      <text variable="page" prefix=", pp. " suffix="."/>
    </layout>
  </bibliography>
</style>`
  }

  private getIEEEStyle(): string {
    return `<?xml version="1.0" encoding="utf-8"?>
<style xmlns="http://purl.org/net/xbiblio/csl" class="in-text" version="1.0">
  <info><title>IEEE</title><id>ieee</id></info>
  <citation>
    <layout prefix="[" suffix="]" delimiter=",">
      <text variable="citation-number"/>
    </layout>
  </citation>
  <bibliography second-field-align="flush">
    <layout>
      <text variable="citation-number" prefix="[" suffix="] "/>
      <text variable="author" suffix=", "/>
      <text variable="title" suffix=", " quotes="true"/>
      <text variable="container-title" suffix=", " font-style="italic"/>
      <text variable="volume" suffix=", "/>
      <text variable="page" prefix="pp. " suffix=", "/>
      <date variable="issued" prefix=" "><date-part name="year"/></date>
    </layout>
  </bibliography>
</style>`
  }

  private getGBT7714Style(): string {
    return `<?xml version="1.0" encoding="utf-8"?>
<style xmlns="http://purl.org/net/xbiblio/csl" class="in-text" version="1.0" default-locale="zh-CN">
  <info><title>GB/T 7714-2015</title><id>gb-t-7714-2015</id></info>
  <citation et-al-min="3" et-al-use-first="3">
    <layout prefix="[" suffix="]" delimiter=",">
      <text variable="citation-number"/>
    </layout>
  </citation>
  <bibliography second-field-align="flush">
    <layout>
      <text variable="citation-number" prefix="[" suffix="] "/>
      <text variable="author" suffix=". "/>
      <text variable="title" suffix=". "/>
      <text variable="container-title" suffix=", "/>
      <text variable="volume" suffix="("/>
      <text variable="issue" suffix="): "/>
      <text variable="page" suffix=", "/>
      <date variable="issued"><date-part name="year"/></date>
    </layout>
  </bibliography>
</style>`
  }

  private getZhLocale(): string {
    return `<?xml version="1.0" encoding="utf-8"?>
<locale xmlns="http://purl.org/net/xbiblio/csl" version="1.0" xml:lang="zh-CN">
  <terms>
    <term name="and">和</term>
    <term name="et-al">等</term>
    <term name="in">载</term>
    <term name="page">页</term>
    <term name="accessed">访问于</term>
  </terms>
</locale>`
  }

  private getEnLocale(): string {
    return `<?xml version="1.0" encoding="utf-8"?>
<locale xmlns="http://purl.org/net/xbiblio/csl" version="1.0" xml:lang="en-US">
  <terms>
    <term name="and">and</term>
    <term name="et-al">et al.</term>
    <term name="in">in</term>
    <term name="page">p.</term>
    <term name="accessed">accessed</term>
  </terms>
</locale>`
  }
}

export function referenceToCSLItem(ref: {
  id: string
  title: string
  authors?: Array<{ family?: string; given?: string; literal?: string }>
  year?: number
  source?: string
  volume?: string
  issue?: string
  pages?: string
  doi?: string
  url?: string
  publisher?: string
  type?: string
}): CSL.Item {
  const item: CSL.Item = {
    id: ref.id,
    type: mapTypeToCSL(ref.type || 'article'),
    title: ref.title,
  }

  if (ref.authors && ref.authors.length > 0) {
    item.author = ref.authors.map(a => ({
      family: a.family,
      given: a.given,
      literal: a.literal,
    }))
  }

  if (ref.year) {
    item.issued = { 'date-parts': [[ref.year]] }
  }

  if (ref.source) {
    item['container-title'] = ref.source
  }

  if (ref.volume) item.volume = ref.volume
  if (ref.issue) item.issue = ref.issue
  if (ref.pages) item.page = ref.pages
  if (ref.doi) item.DOI = ref.doi
  if (ref.url) item.URL = ref.url
  if (ref.publisher) item.publisher = ref.publisher

  return item
}

function mapTypeToCSL(type: string): CSL.Item['type'] {
  const typeMap: Record<string, CSL.Item['type']> = {
    'article': 'article-journal',
    'book': 'book',
    'book-chapter': 'chapter',
    'conference': 'paper-conference',
    'thesis': 'thesis',
    'report': 'report',
    'web': 'webpage',
    'preprint': 'article',
  }
  return typeMap[type] || 'document'
}

export const citationFormatter = new CitationFormatter()