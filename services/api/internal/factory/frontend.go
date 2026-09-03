package factory

import (
	"fmt"

	"github.com/cherry/api/internal/store"
)

func frontendFiles(stack store.ProjectStack, name, slug, brief string) ([]fileSpec, error) {
	switch stack {
	case store.StackExpo:
		return expoCleanArch(name, slug, brief), nil
	case store.StackFlutter:
		return flutterCleanArch(name, slug, brief), nil
	case store.StackNative:
		return swiftUICleanArch(name, slug, brief), nil
	default:
		return nil, fmt.Errorf("unhandled stack: %s", stack)
	}
}

func expoCleanArch(name, slug, brief string) []fileSpec {
	title := jsonStr(name)
	desc := jsonStr(brief)
	return []fileSpec{
		{rel: "frontend/ARCHITECTURE.md", kind: "frontend", body: expoArchDoc()},
		{rel: "frontend/package.json", kind: "frontend", body: `{
  "name": "` + slug + `",
  "private": true,
  "main": "expo-router/entry",
  "scripts": {
    "start": "expo start",
    "android": "expo start --android",
    "ios": "expo start --ios"
  },
  "dependencies": {
    "expo": "~57.0.0",
    "expo-router": "~6.0.0",
    "expo-status-bar": "~3.0.0",
    "react": "19.2.0",
    "react-native": "0.86.3",
    "react-native-safe-area-context": "~5.6.0",
    "react-native-screens": "~4.16.0"
  },
  "devDependencies": {
    "@types/react": "~19.2.0",
    "typescript": "~5.9.0"
  }
}
`},
		{rel: "frontend/app.json", kind: "frontend", body: `{
  "expo": {
    "name": ` + title + `,
    "slug": "` + slug + `",
    "scheme": "` + slug + `",
    "newArchEnabled": true,
    "plugins": ["expo-router"]
  }
}
`},
		{rel: "frontend/tsconfig.json", kind: "frontend", body: `{
  "extends": "expo/tsconfig.base",
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitOverride": true
  },
  "include": ["**/*.ts", "**/*.tsx"]
}
`},
		{rel: "frontend/src/domain/entities/item.ts", kind: "frontend", body: `export type Item = {
  id: string;
  title: string;
  subtitle: string;
};
`},
		{rel: "frontend/src/domain/repositories/item-repository.ts", kind: "frontend", body: `import type { Item } from "../entities/item";

export interface ItemRepository {
  list(): Promise<readonly Item[]>;
}
`},
		{rel: "frontend/src/domain/usecases/list-items.ts", kind: "frontend", body: `import type { Item } from "../entities/item";
import type { ItemRepository } from "../repositories/item-repository";

export class ListItems {
  constructor(private readonly repository: ItemRepository) {}

  execute(): Promise<readonly Item[]> {
    return this.repository.list();
  }
}
`},
		{rel: "frontend/src/data/sources/item-memory-source.ts", kind: "frontend", body: `import type { Item } from "../../domain/entities/item";

export class ItemMemorySource {
  constructor(
    private readonly title: string,
    private readonly subtitle: string,
  ) {}

  async list(): Promise<readonly Item[]> {
    return [
      { id: "home", title: this.title, subtitle: this.subtitle },
    ];
  }
}
`},
		{rel: "frontend/src/data/repositories/item-repository-impl.ts", kind: "frontend", body: `import type { Item } from "../../domain/entities/item";
import type { ItemRepository } from "../../domain/repositories/item-repository";
import type { ItemMemorySource } from "../sources/item-memory-source";

export class ItemRepositoryImpl implements ItemRepository {
  constructor(private readonly source: ItemMemorySource) {}

  list(): Promise<readonly Item[]> {
    return this.source.list();
  }
}
`},
		{rel: "frontend/src/app/composition.ts", kind: "frontend", body: `import { ItemMemorySource } from "../data/sources/item-memory-source";
import { ItemRepositoryImpl } from "../data/repositories/item-repository-impl";
import { ListItems } from "../domain/usecases/list-items";

export function composeListItems(title: string, subtitle: string): ListItems {
  const source = new ItemMemorySource(title, subtitle);
  const repository = new ItemRepositoryImpl(source);
  return new ListItems(repository);
}
`},
		{rel: "frontend/src/presentation/hooks/use-home.ts", kind: "frontend", body: `import { useEffect, useState } from "react";

import { composeListItems } from "../../app/composition";
import type { Item } from "../../domain/entities/item";

type HomeState =
  | { status: "loading" }
  | { status: "ready"; items: readonly Item[] }
  | { status: "error"; message: string };

export function useHome(title: string, subtitle: string): HomeState {
  const [state, setState] = useState<HomeState>({ status: "loading" });

  useEffect(() => {
    let cancelled = false;
    composeListItems(title, subtitle)
      .execute()
      .then((items) => {
        if (!cancelled) {
          setState({ status: "ready", items });
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setState({
            status: "error",
            message: err instanceof Error ? err.message : "Yüklenemedi",
          });
        }
      });
    return () => {
      cancelled = true;
    };
  }, [title, subtitle]);

  return state;
}
`},
		{rel: "frontend/src/presentation/screens/login-screen.tsx", kind: "frontend", body: `import { Pressable, Text, TextInput, View } from "react-native";
import { router } from "expo-router";

type Props = {
  appName: string;
};

export function LoginScreen({ appName }: Props) {
  return (
    <View style={{ flex: 1, justifyContent: "center", padding: 24, gap: 12, backgroundColor: "#0E1114" }}>
      <Text style={{ fontSize: 12, letterSpacing: 1, color: "#8B939C" }}>{appName}</Text>
      <Text style={{ fontSize: 22, fontWeight: "600", color: "#E8E4DC" }}>Giriş</Text>
      <Text style={{ color: "#8B939C" }}>E-posta ve şifre ile içeri. SMS yok.</Text>
      <TextInput
        accessibilityLabel="E-posta"
        placeholder="e-posta"
        placeholderTextColor="#8B939C"
        autoCapitalize="none"
        keyboardType="email-address"
        style={{
          borderWidth: 1,
          borderColor: "#2A323A",
          borderRadius: 8,
          paddingHorizontal: 12,
          paddingVertical: 10,
          color: "#E8E4DC",
          backgroundColor: "#161B20",
        }}
      />
      <TextInput
        accessibilityLabel="Şifre"
        placeholder="şifre"
        placeholderTextColor="#8B939C"
        secureTextEntry
        style={{
          borderWidth: 1,
          borderColor: "#2A323A",
          borderRadius: 8,
          paddingHorizontal: 12,
          paddingVertical: 10,
          color: "#E8E4DC",
          backgroundColor: "#161B20",
        }}
      />
      <Pressable
        accessibilityRole="button"
        onPress={() => {
          router.replace("/home");
        }}
        style={{
          marginTop: 8,
          backgroundColor: "#C4A574",
          borderRadius: 8,
          paddingVertical: 12,
          alignItems: "center",
        }}
      >
        <Text style={{ color: "#0E1114", fontWeight: "600" }}>Devam</Text>
      </Pressable>
    </View>
  );
}
`},
		{rel: "frontend/src/presentation/screens/home-screen.tsx", kind: "frontend", body: `import { ActivityIndicator, Text, View } from "react-native";

import { useHome } from "../hooks/use-home";

type Props = {
  title: string;
  subtitle: string;
};

export function HomeScreen({ title, subtitle }: Props) {
  const state = useHome(title, subtitle);

  if (state.status === "loading") {
    return (
      <View style={{ flex: 1, justifyContent: "center", padding: 24, backgroundColor: "#0E1114" }}>
        <ActivityIndicator color="#C4A574" />
      </View>
    );
  }
  if (state.status === "error") {
    return (
      <View style={{ flex: 1, justifyContent: "center", padding: 24, backgroundColor: "#0E1114" }}>
        <Text style={{ color: "#E8E4DC" }}>{state.message}</Text>
      </View>
    );
  }
  const item = state.items[0];
  return (
    <View style={{ flex: 1, justifyContent: "center", padding: 24, gap: 8, backgroundColor: "#0E1114" }}>
      <Text style={{ fontSize: 22, fontWeight: "600", color: "#E8E4DC" }}>{item?.title ?? title}</Text>
      <Text style={{ color: "#8B939C" }}>{item?.subtitle ?? subtitle}</Text>
      <Text style={{ marginTop: 12, color: "#8B939C" }}>Bugünkü siparişler burada. Henüz kayıt yok.</Text>
    </View>
  );
}
`},
		{rel: "frontend/app/_layout.tsx", kind: "frontend", body: `import { Stack } from "expo-router";

export default function RootLayout() {
  return <Stack screenOptions={{ headerShown: false }} />;
}
`},
		{rel: "frontend/app/index.tsx", kind: "frontend", body: `import { LoginScreen } from "../src/presentation/screens/login-screen";

export default function Index() {
  return <LoginScreen appName={` + title + `} />;
}
`},
		{rel: "frontend/app/home.tsx", kind: "frontend", body: `import { HomeScreen } from "../src/presentation/screens/home-screen";

export default function HomeRoute() {
  return (
    <HomeScreen title={` + title + `} subtitle={` + desc + `} />
  );
}
`},
	}
}

func expoArchDoc() string {
	return "# Clean Architecture — Expo SDK 57\n\n" +
		"Katmanlar / Layers:\n\n" +
		"- src/domain — varlıklar, repository port, use case. React yok.\n" +
		"- src/data — kaynak ve repository impl.\n" +
		"- src/presentation — ekran ve hook (function component, hook; class yok).\n" +
		"- src/app — composition root.\n" +
		"- app/ — Expo Router v6 giriş. İş kuralı burada durmaz.\n\n" +
		"Dil: TypeScript strict, React 19.2, React Native 0.86, Expo SDK 57, New Architecture açık.\n" +
		"HTML site yazma. Katmanları düz App.js içine yığma.\n"
}

func flutterArchDoc() string {
	return "# Clean Architecture — Flutter 3.47 / Dart 3.13\n\n" +
		"Özellik klasörü / Feature folders:\n\n" +
		"- lib/features/<feature>/domain — entity, repository port, use case. Flutter import yok.\n" +
		"- lib/features/<feature>/data — data source + repository impl.\n" +
		"- lib/features/<feature>/presentation — page / view.\n" +
		"- lib/core — Failure, UseCase.\n" +
		"- lib/app — composition root + Material 3 teması.\n\n" +
		"Dil: Dart 3.13 (sealed class, final class, abstract interface class, super parameters, const).\n" +
		"print yok. Tek main.dart içine tüm uygulamayı yığma. HTML yazma.\n"
}

func swiftArchDoc() string {
	return "# Clean Architecture — SwiftUI\n\n" +
		"Katmanlar / Layers:\n\n" +
		"- Domain — entity, repository protocol, use case. SwiftUI import yok.\n" +
		"- Data — source + repository impl.\n" +
		"- Presentation — SwiftUI View + @Observable ViewModel.\n" +
		"- App — @main + composition root.\n\n" +
		"Dil: Swift 6, SwiftUI (UIKit ekran yok), iOS 18+, async/await, @Observable, #Preview, Sendable.\n" +
		"Tek ContentView.swift içine tüm katmanları yığma. HTML yazma.\n"
}

func flutterCleanArch(name, slug, brief string) []fileSpec {
	pkg := dartPackage(slug)
	title := dartStr(name)
	desc := dartStr(brief)
	return []fileSpec{
		{rel: "frontend/ARCHITECTURE.md", kind: "frontend", body: flutterArchDoc()},
		{rel: "frontend/pubspec.yaml", kind: "frontend", body: `name: ` + pkg + `
description: ` + yamlStr(brief) + `
publish_to: "none"
version: 0.1.0+1

environment:
  sdk: ">=3.13.0 <4.0.0"

dependencies:
  flutter:
    sdk: flutter

dev_dependencies:
  flutter_test:
    sdk: flutter
  flutter_lints: ^6.0.0

flutter:
  uses-material-design: true
`},
		{rel: "frontend/analysis_options.yaml", kind: "frontend", body: `include: package:flutter_lints/flutter.yaml

linter:
  rules:
    prefer_const_constructors: true
    prefer_const_literals_to_create_immutables: true
    avoid_print: true
    use_super_parameters: true
`},
		{rel: "frontend/lib/core/error/failure.dart", kind: "frontend", body: `sealed class Failure {
  const Failure(this.message);
  final String message;
}

final class LoadFailure extends Failure {
  const LoadFailure(super.message);
}
`},
		{rel: "frontend/lib/core/usecase/usecase.dart", kind: "frontend", body: `abstract interface class UseCase<T> {
  Future<T> call();
}
`},
		{rel: "frontend/lib/features/home/domain/entities/home_item.dart", kind: "frontend", body: `final class HomeItem {
  const HomeItem({required this.id, required this.title, required this.subtitle});

  final String id;
  final String title;
  final String subtitle;
}
`},
		{rel: "frontend/lib/features/home/domain/repositories/home_repository.dart", kind: "frontend", body: `import '../entities/home_item.dart';

abstract interface class HomeRepository {
  Future<List<HomeItem>> list();
}
`},
		{rel: "frontend/lib/features/home/domain/usecases/list_home_items.dart", kind: "frontend", body: `import '../../../../core/usecase/usecase.dart';
import '../entities/home_item.dart';
import '../repositories/home_repository.dart';

final class ListHomeItems implements UseCase<List<HomeItem>> {
  const ListHomeItems(this._repository);

  final HomeRepository _repository;

  @override
  Future<List<HomeItem>> call() => _repository.list();
}
`},
		{rel: "frontend/lib/features/home/data/datasources/home_local_data_source.dart", kind: "frontend", body: `import '../../domain/entities/home_item.dart';

class HomeLocalDataSource {
  const HomeLocalDataSource({required this.title, required this.subtitle});

  final String title;
  final String subtitle;

  Future<List<HomeItem>> list() async {
    return [
      HomeItem(id: 'home', title: title, subtitle: subtitle),
    ];
  }
}
`},
		{rel: "frontend/lib/features/home/data/repositories/home_repository_impl.dart", kind: "frontend", body: `import '../../domain/entities/home_item.dart';
import '../../domain/repositories/home_repository.dart';
import '../datasources/home_local_data_source.dart';

final class HomeRepositoryImpl implements HomeRepository {
  const HomeRepositoryImpl(this._source);

  final HomeLocalDataSource _source;

  @override
  Future<List<HomeItem>> list() => _source.list();
}
`},
		{rel: "frontend/lib/features/home/presentation/pages/home_page.dart", kind: "frontend", body: `import 'package:flutter/material.dart';

import '../../domain/entities/home_item.dart';
import '../../domain/usecases/list_home_items.dart';

class HomePage extends StatelessWidget {
  const HomePage({super.key, required this.listHomeItems});

  final ListHomeItems listHomeItems;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: FutureBuilder<List<HomeItem>>(
        future: listHomeItems(),
        builder: (context, snapshot) {
          if (snapshot.hasError) {
            return Center(child: Text('${snapshot.error}'));
          }
          if (!snapshot.hasData) {
            return const Center(child: CircularProgressIndicator());
          }
          final item = snapshot.data!.first;
          return Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(item.title, style: Theme.of(context).textTheme.headlineMedium),
                const SizedBox(height: 8),
                Text(item.subtitle),
              ],
            ),
          );
        },
      ),
    );
  }
}
`},
		{rel: "frontend/lib/app/di.dart", kind: "frontend", body: `import '../features/home/data/datasources/home_local_data_source.dart';
import '../features/home/data/repositories/home_repository_impl.dart';
import '../features/home/domain/usecases/list_home_items.dart';

ListHomeItems composeHome({required String title, required String subtitle}) {
  final source = HomeLocalDataSource(title: title, subtitle: subtitle);
  final repository = HomeRepositoryImpl(source);
  return ListHomeItems(repository);
}
`},
		{rel: "frontend/lib/app/app.dart", kind: "frontend", body: `import 'package:flutter/material.dart';

import '../features/home/presentation/pages/home_page.dart';
import 'di.dart';

class App extends StatelessWidget {
  const App({super.key, required this.title, required this.subtitle});

  final String title;
  final String subtitle;

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: title,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: const Color(0xFFC4A574)),
        useMaterial3: true,
      ),
      home: HomePage(listHomeItems: composeHome(title: title, subtitle: subtitle)),
    );
  }
}
`},
		{rel: "frontend/lib/main.dart", kind: "frontend", body: `import 'package:flutter/material.dart';

import 'package:` + pkg + `/app/app.dart';

void main() {
  runApp(const App(title: '` + title + `', subtitle: '` + desc + `'));
}
`},
	}
}

func swiftUICleanArch(name, slug, brief string) []fileSpec {
	app := pascalIdent(slug) + "App"
	title := swiftStr(name)
	desc := swiftStr(brief)
	return []fileSpec{
		{rel: "frontend/ARCHITECTURE.md", kind: "frontend", body: swiftArchDoc()},
		{rel: "frontend/README.md", kind: "frontend", body: "# " + name + "\n\nSwiftUI (iOS 18+, Swift 6) Clean Architecture.\nXcode 16+ ile `App/` hedefini aç. UIKit ViewController yazma.\n"},
		{rel: "frontend/Domain/Entities/HomeItem.swift", kind: "frontend", body: `import Foundation

struct HomeItem: Identifiable, Equatable, Sendable {
    let id: String
    let title: String
    let subtitle: String
}
`},
		{rel: "frontend/Domain/Repositories/HomeRepository.swift", kind: "frontend", body: `protocol HomeRepository: Sendable {
    func list() async throws -> [HomeItem]
}
`},
		{rel: "frontend/Domain/UseCases/ListHomeItems.swift", kind: "frontend", body: `struct ListHomeItems: Sendable {
    private let repository: HomeRepository

    init(repository: HomeRepository) {
        self.repository = repository
    }

    func callAsFunction() async throws -> [HomeItem] {
        try await repository.list()
    }
}
`},
		{rel: "frontend/Data/Sources/HomeMemorySource.swift", kind: "frontend", body: `struct HomeMemorySource: Sendable {
    let title: String
    let subtitle: String

    func list() async throws -> [HomeItem] {
        [
            HomeItem(id: "home", title: title, subtitle: subtitle),
        ]
    }
}
`},
		{rel: "frontend/Data/Repositories/HomeRepositoryImpl.swift", kind: "frontend", body: `struct HomeRepositoryImpl: HomeRepository {
    private let source: HomeMemorySource

    init(source: HomeMemorySource) {
        self.source = source
    }

    func list() async throws -> [HomeItem] {
        try await source.list()
    }
}
`},
		{rel: "frontend/Presentation/Home/HomeViewModel.swift", kind: "frontend", body: `import Foundation
import Observation

enum HomeLoadState: Equatable {
    case loading
    case ready([HomeItem])
    case failed(String)
}

@MainActor
@Observable
final class HomeViewModel {
    private(set) var state: HomeLoadState = .loading
    private let listHomeItems: ListHomeItems

    init(listHomeItems: ListHomeItems) {
        self.listHomeItems = listHomeItems
    }

    func load() async {
        state = .loading
        do {
            state = .ready(try await listHomeItems())
        } catch {
            state = .failed(error.localizedDescription)
        }
    }
}
`},
		{rel: "frontend/Presentation/Home/HomeView.swift", kind: "frontend", body: `import SwiftUI

struct HomeView: View {
    @State private var model: HomeViewModel

    init(model: HomeViewModel) {
        _model = State(initialValue: model)
    }

    var body: some View {
        NavigationStack {
            Group {
                switch model.state {
                case .loading:
                    ProgressView()
                case .failed(let message):
                    Text(message)
                case .ready(let items):
                    if let item = items.first {
                        VStack(alignment: .leading, spacing: 8) {
                            Text(item.title)
                                .font(.title)
                            Text(item.subtitle)
                                .foregroundStyle(.secondary)
                        }
                        .padding(24)
                    }
                }
            }
            .navigationTitle("Ana ekran")
            .task { await model.load() }
        }
    }
}

#Preview {
    HomeView(model: Composition.homeModel(title: "` + title + `", subtitle: "` + desc + `"))
}
`},
		{rel: "frontend/App/Composition.swift", kind: "frontend", body: `enum Composition {
    @MainActor
    static func homeModel(title: String, subtitle: String) -> HomeViewModel {
        let source = HomeMemorySource(title: title, subtitle: subtitle)
        let repository = HomeRepositoryImpl(source: source)
        return HomeViewModel(listHomeItems: ListHomeItems(repository: repository))
    }
}
`},
		{rel: "frontend/App/" + app + ".swift", kind: "frontend", body: `import SwiftUI

@main
struct ` + app + `: App {
    var body: some Scene {
        WindowGroup {
            HomeView(
                model: Composition.homeModel(
                    title: "` + title + `",
                    subtitle: "` + desc + `"
                )
            )
        }
    }
}
`},
	}
}

