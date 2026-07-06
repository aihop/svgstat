function createEmptyProjectStats() {
    return {
        projectId: '',
        date: '',
        pv: 0,
        uv: 0,
        requests: 0,
        bots: 0,
        referrers: {},
        countries: {},
        regions: {},
        cities: {},
        devices: {},
        browsers: {},
        paths: {},
        ips: {},
        visitors: []
    };
}

// Define the Alpine component
function spaApp() {
    return {
        lang: localStorage.getItem('svgstat-lang') || 'en',
        user: null,
        currentPage: 'home',
        projects: [],
        loading: false,
        creating: false,
        showCreateModal: false,
        showCodeModal: false,
        expandedVisitorId: null,
        selectedProject: null,
        projectStats: createEmptyProjectStats(),
        loadingStats: false,
        codeSettings: {
            counterName: 'visits',
            counterLabel: 'Visits',
            counterColor: '',
            badgeName: 'downloads',
            badgeLabel: 'Downloads',
            badgeColor: '',
            badgeStyle: ''
        },
        newProject: {
            name: '',
            slug: '',
            description: ''
        },
        loginForm: {
            email: '',
            password: ''
        },
        registerForm: {
            name: '',
            email: '',
            password: ''
        },
        loginLoading: false,
        registerLoading: false,
        loginError: null,
        registerError: null,
        registerSuccess: null,

        t(key) {
            return translations[this.lang]?.[key] || translations.en[key] || key;
        },

        init() {
            this.$watch('lang', (val) => {
                localStorage.setItem('svgstat-lang', val);
                document.documentElement.lang = val;
            });
            this.$watch('selectedProject', (val) => {
                if (val && this.currentPage === 'project-detail') {
                    this.loadStats(val.id);
                }
            });
            document.documentElement.lang = this.lang;
            this.parseRoute();
            window.addEventListener('popstate', () => this.parseRoute());
            this.checkAuth();
            if (this.currentPage === 'dashboard' || this.currentPage === 'project-detail') {
                this.loadProjects();
            }
        },

        parseRoute() {
            const path = window.location.pathname;
            if (path === '/login') {
                this.currentPage = 'login';
                this.selectedProject = null;
                this.showCodeModal = false;
                this.expandedVisitorId = null;
                this.projectStats = createEmptyProjectStats();
            } else if (path === '/register') {
                this.currentPage = 'register';
                this.selectedProject = null;
                this.showCodeModal = false;
                this.expandedVisitorId = null;
                this.projectStats = createEmptyProjectStats();
            } else if (path === '/dashboard') {
                this.currentPage = 'dashboard';
                this.selectedProject = null;
                this.showCodeModal = false;
                this.expandedVisitorId = null;
                this.projectStats = createEmptyProjectStats();
            } else if (path.startsWith('/dashboard/')) {
                // 处理项目详情路由
                const slug = path.substring('/dashboard/'.length);
                this.currentPage = 'project-detail';
                this.showCodeModal = false;
                this.expandedVisitorId = null;
                // 如果项目已经加载过，则查找对应的项目
                if (this.projects.length > 0) {
                    const project = this.projects.find(p => p.slug === slug);
                    if (project) {
                        this.selectedProject = project;
                    }
                }
            } else {
                this.currentPage = 'home';
                this.selectedProject = null;
                this.showCodeModal = false;
                this.expandedVisitorId = null;
                this.projectStats = createEmptyProjectStats();
            }
        },

        navigate(path) {
            window.history.pushState({}, '', path);
            this.parseRoute();
            window.scrollTo(0, 0);
            if (this.currentPage === 'dashboard' || this.currentPage === 'project-detail') {
                this.loadProjects();
            }
        },

        async checkAuth() {
            try {
                const res = await fetch('/api/v1/auth/me', { credentials: 'same-origin' });
                const data = await res.json();
                if (data.success) {
                    this.user = data.data;
                }
            } catch (e) {
                console.error('Auth check failed', e);
            }
        },

        async login() {
            this.loginLoading = true;
            this.loginError = null;

            try {
                const res = await fetch('/api/v1/auth/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(this.loginForm),
                    credentials: 'same-origin'
                });

                const data = await res.json();

                if (data.success) {
                    this.user = data.data.user;
                    this.loginForm = { email: '', password: '' };
                    this.navigate('/dashboard');
                } else {
                    this.loginError = data.error || (this.lang === 'zh' ? '登录失败' : 'Invalid credentials');
                }
            } catch (e) {
                this.loginError = this.lang === 'zh' ? '发生错误，请重试' : 'Something went wrong. Please try again.';
                console.error(e);
            } finally {
                this.loginLoading = false;
            }
        },

        async register() {
            this.registerLoading = true;
            this.registerError = null;
            this.registerSuccess = null;

            try {
                const res = await fetch('/api/v1/auth/register', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(this.registerForm),
                    credentials: 'same-origin'
                });

                const data = await res.json();

                if (data.success) {
                    this.user = data.data.user;
                    this.registerForm = { name: '', email: '', password: '' };
                    this.registerSuccess = this.lang === 'zh' ? '账户创建成功！正在跳转...' : 'Account created! Redirecting...';
                    setTimeout(() => {
                        this.navigate('/dashboard');
                    }, 1000);
                } else {
                    this.registerError = data.error || (this.lang === 'zh' ? '发生错误' : 'Something went wrong');
                }
            } catch (e) {
                this.registerError = this.lang === 'zh' ? '发生错误，请重试' : 'Something went wrong. Please try again.';
                console.error(e);
            } finally {
                this.registerLoading = false;
            }
        },

        async logout() {
            try {
                await fetch('/api/v1/auth/logout', {
                    method: 'POST',
                    credentials: 'same-origin'
                });
                this.user = null;
                this.navigate('/');
            } catch (e) {
                console.error('Logout failed', e);
            }
        },

        async loadProjects() {
            this.loading = true;
            try {
                const res = await fetch('/api/v1/projects', { credentials: 'same-origin' });
                const data = await res.json();
                if (data.success) {
                    this.projects = data.data;
                    // 如果当前是项目详情页面，查找对应的项目
                    if (this.currentPage === 'project-detail' && !this.selectedProject) {
                        const path = window.location.pathname;
                        const slug = path.substring('/dashboard/'.length);
                        const project = this.projects.find(p => p.slug === slug);
                        if (project) {
                            this.selectedProject = project;
                        }
                    }
                }
            } catch (e) {
                console.error('Failed to load projects', e);
            } finally {
                this.loading = false;
            }
        },

        async createProject() {
            this.creating = true;
            try {
                const res = await fetch('/api/v1/projects', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(this.newProject),
                    credentials: 'same-origin'
                });

                const data = await res.json();
                if (data.success) {
                    this.showCreateModal = false;
                    this.newProject = { name: '', slug: '', description: '' };
                    await this.loadProjects();
                } else {
                    alert(data.error || this.t('errorCreate'));
                }
            } catch (e) {
                console.error(e);
                alert(this.t('errorGeneric'));
            } finally {
                this.creating = false;
            }
        },

        async deleteProject(id) {
            if (!confirm(this.t('confirmDelete'))) return;

            try {
                const res = await fetch(`/api/v1/projects/${id}`, {
                    method: 'DELETE',
                    credentials: 'same-origin'
                });

                if (res.ok) {
                    await this.loadProjects();
                }
            } catch (e) {
                console.error(e);
            }
        },

        viewProjectStats(project) {
            this.navigate(`/dashboard/${project.slug}`);
        },

        resetCodeSettings() {
            this.codeSettings = {
                counterName: 'visits',
                counterLabel: this.lang === 'zh' ? '访问量' : 'Visits',
                counterColor: '',
                badgeName: 'downloads',
                badgeLabel: this.lang === 'zh' ? '下载量' : 'Downloads',
                badgeColor: '',
                badgeStyle: ''
            };
        },

        openCodeModal(project) {
            this.selectedProject = project;
            this.resetCodeSettings();
            this.showCodeModal = true;
        },

        closeCodeModal() {
            this.showCodeModal = false;
            if (this.currentPage !== 'project-detail') {
                this.selectedProject = null;
            }
        },

        getQueryString(params) {
            const search = new URLSearchParams();
            Object.entries(params || {}).forEach(([key, value]) => {
                if (value !== null && value !== undefined && String(value).trim() !== '') {
                    search.set(key, String(value).trim());
                }
            });
            const result = search.toString();
            return result ? `?${result}` : '';
        },

        getCounterSvgPath(project = this.selectedProject) {
            if (!project) return '';
            const name = this.codeSettings.counterName || 'visits';
            const query = this.getQueryString({
                label: this.codeSettings.counterLabel,
                color: this.codeSettings.counterColor
            });
            return `/svg/${project.slug}/counter/${name}.svg${query}`;
        },

        getBadgeSvgPath(project = this.selectedProject) {
            if (!project) return '';
            const name = this.codeSettings.badgeName || 'downloads';
            const query = this.getQueryString({
                label: this.codeSettings.badgeLabel,
                color: this.codeSettings.badgeColor,
                style: this.codeSettings.badgeStyle
            });
            return `/svg/${project.slug}/badge/${name}.svg${query}`;
        },

        getCounterMarkdown(project = this.selectedProject) {
            const label = this.codeSettings.counterLabel || this.codeSettings.counterName || 'Visits';
            const path = this.getCounterSvgPath(project);
            return path ? `![${label}](${path})` : '';
        },

        getBadgeMarkdown(project = this.selectedProject) {
            const label = this.codeSettings.badgeLabel || this.codeSettings.badgeName || 'Downloads';
            const path = this.getBadgeSvgPath(project);
            return path ? `![${label}](${path})` : '';
        },

        copyText(text) {
            navigator.clipboard.writeText(window.location.origin + text).then(() => {
                alert(this.t('copied'));
            });
        },

        getSortedEntries(record, limit = null) {
            const entries = Object.entries(record || {}).sort((a, b) => b[1] - a[1]);
            return limit ? entries.slice(0, limit) : entries;
        },

        getBarStyle(count, record) {
            const values = Object.values(record || {});
            const max = values.length ? Math.max(...values) : 0;
            const width = max > 0 ? (count / max) * 100 : 0;
            return `width: ${width}%`;
        },

        shortVisitorId(visitorId) {
            if (!visitorId) return '-';
            return visitorId.length > 12 ? `${visitorId.slice(0, 12)}...` : visitorId;
        },

        formatDateTime(value) {
            if (!value) return '-';
            const date = new Date(value);
            if (Number.isNaN(date.getTime())) return '-';
            return date.toLocaleString(this.lang === 'zh' ? 'zh-CN' : 'en-US');
        },

        formatVisitorLocation(visitor) {
            const parts = [visitor.country, visitor.region, visitor.city].filter(Boolean);
            return parts.length ? parts.join(' / ') : '-';
        },

        getVisibleVisitors() {
            return (this.projectStats.visitors || []).slice(0, 20);
        },

        toggleVisitorDetail(visitorId) {
            this.expandedVisitorId = this.expandedVisitorId === visitorId ? null : visitorId;
        },
        
        async loadStats(projectId) {
            this.loadingStats = true;
            this.expandedVisitorId = null;
            this.projectStats = createEmptyProjectStats();
            try {
                const res = await fetch(`/api/v1/projects/${projectId}/stats`, { credentials: 'same-origin' });
                const data = await res.json();
                if (data.success) {
                    this.projectStats = {
                        ...createEmptyProjectStats(),
                        ...data.data
                    };
                }
            } catch (e) {
                console.error('Failed to load stats', e);
            } finally {
                this.loadingStats = false;
            }
        }
    };
}

// Load all components after DOM is ready
async function loadComponents() {
	// Function to load and insert a single component
	async function loadAndInsert(path, containerId) {
		try {
			const response = await fetch(path);
			const html = await response.text();
			const container = document.getElementById(containerId);
			if (container) {
				container.innerHTML = html;
			}
		} catch (e) {
			console.error('Failed to load component:', path, e);
		}
	}

	// Load navbar
	await loadAndInsert('/components/Navbar.html', 'navbar-container');

	// Load all pages
	const [home, login, register, dashboard, dashboardProject] = await Promise.all([
		fetch('/components/pages/Home.html').then(r => r.text()),
		fetch('/components/pages/Login.html').then(r => r.text()),
		fetch('/components/pages/Register.html').then(r => r.text()),
		fetch('/components/pages/Dashboard.html').then(r => r.text()),
		fetch('/components/pages/DashboardProject.html').then(r => r.text())
	]);
	const pageContainer = document.getElementById('page-container');
	if (pageContainer) {
		pageContainer.innerHTML = home + login + register + dashboard + dashboardProject;
	}

	// Load all modals
	const [createModal, detailModal] = await Promise.all([
		fetch('/components/CreateProjectModal.html').then(r => r.text()),
		fetch('/components/ProjectDetailModal.html').then(r => r.text())
	]);
	const modalsContainer = document.getElementById('modals-container');
	if (modalsContainer) {
		modalsContainer.innerHTML = createModal + detailModal;
	}
}

// Start loading components when DOM is ready
document.addEventListener('DOMContentLoaded', loadComponents);
